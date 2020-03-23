from aws_fuzzy.cli import pass_environment, query
from .common import common_params

import click
import re
import os
import subprocess

from iterfzf import iterfzf
from os.path import expanduser


@click.command("ssh")
@common_params(p=False)
@click.option(
    '-u',
    '--user',
    default="''",
    show_default=True,
    help='Username to use with SSH')
@click.option(
    '-k', '--key', default="''", show_default=True, help='SSH key path')
@pass_environment
def cli(ctx, **kwargs):
    """SSH to EC2 instance"""

    kwargs['service'] = "AWS::EC2::Instance"
    kwargs[
        'select'] = "resourceId, accountId, configuration.privateIpAddress, tags"
    f = f"resourceType like '{kwargs['service']}' AND " \
         "configuration.state.name like 'running'"

    if kwargs['filter'] != "''":
        kwargs['filter'] = f"{f}' AND {kwargs['filter']}"
    else:
        kwargs['filter'] = f"{f}"

    ret = query(ctx, **kwargs)
    ctx.vlog(f"Return form query function: {ret}")
    out = []
    for i in ret:
        name = "<unnamed>"
        tags = []
        for t in i["tags"]:  # search for tag with key "Name"
            tags.append(t['tag'])
            if t["key"] == "Name":
                name = t["value"]
        out.append(
            f'{name}\t{i["configuration"]["privateIpAddress"]}\t{i["accountId"]}\t{tags}'
        )

    sel = iterfzf(out)

    if sel is None:
        return

    name, ip, account, tags = sel.split('\t')

    if kwargs['key'] != "''":
        key = f"-i {kwargs['key']}"
    else:
        key = ''

    if kwargs['user'] != "''":
        user = f"-l {kwargs['user']}"
    else:
        user = ''

    ssh_command = f"ssh {key} {user} {ip}"

    ctx.vlog(f"Executing: {ssh_command}")

    subprocess.call(
        ssh_command, shell=True, executable=os.getenv('SHELL', '/bin/bash'))
