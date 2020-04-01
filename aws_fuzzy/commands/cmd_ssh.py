from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params
from .common import cache_params
from .common import get_profile
from .common import check_expired

import click
import re
import os
import subprocess
import shelve

from iterfzf import iterfzf
from os.path import expanduser
from datetime import datetime
from datetime import timedelta


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
@cache_params()
@pass_environment
def cli(ctx, **kwargs):
    """SSH to EC2 instance"""

    if kwargs['account'] == 'all':
        profile = 'all'
    else:
        profile = get_profile(kwargs['account'])
        ctx.vlog(profile)
        if profile != None:
            kwargs['account'] = profile['sso_account_id']
            profile = profile['name']
        else:
            # Error check
            pass

    if kwargs['cache']:
        s = shelve.open(ctx.cache_dir + "/ssh")

        ctx.vlog(profile)
        if profile in s:
            tmp = s[profile]

            if check_expired(tmp['date'], kwargs['cache_time']):
                ret = do_query(ctx, kwargs)
                tmp['instances'] = ret
                tmp['date'] = datetime.utcnow()
            else:
                ret = tmp['instances']
        else:
            ret = do_query(ctx, kwargs)
            s[profile] = {'instances': ret, 'date': datetime.utcnow()}
    else:
        ret = do_query(ctx, kwargs)

    do_fzf(ctx, kwargs, ret)


def do_query(ctx, kwargs):
    ctx.vlog(kwargs)
    kwargs['service'] = "AWS::EC2::Instance"
    kwargs[
        'select'] = "resourceId, accountId, configuration.privateIpAddress, tags"
    f = f"resourceType like '{kwargs['service']}' AND " \
         "configuration.state.name like 'running'"

    if kwargs['filter'] != "''":
        kwargs['filter'] = f"{f}' AND {kwargs['filter']}"
    else:
        kwargs['filter'] = f"{f}"

    if kwargs['account'] != 'all':
        kwargs['filter'] += f" AND accountId like '{kwargs['account']}'"
    if kwargs['region'] != 'all':
        kwargs['filter'] += f" AND awsRegion like '{kwargs['account']}'"

    ret = query(ctx, kwargs)

    if kwargs['pager']:
        return

    ctx.vlog("Return form query function:")
    #ctx.vlog(ret)
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

    return out


def do_fzf(ctx, kwargs, instances):
    sel = iterfzf(instances)

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

    ctx.vlog("Executing:")
    ctx.vlog(ssh_command)

    subprocess.call(
        ssh_command, shell=True, executable=os.getenv('SHELL', '/bin/bash'))
