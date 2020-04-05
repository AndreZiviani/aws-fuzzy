from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import Query
from aws_fuzzy import common

import click
import os
import subprocess

from iterfzf import iterfzf
from datetime import datetime
from datetime import timedelta


class SSH(common.Cache):
    def __init__(self,
                 ctx,
                 Cache,
                 Account,
                 Key=None,
                 User=None,
                 Cache_time=3600):
        super().__init__(ctx, "ssh", Cache_time)

        self.cache = Cache

        self.key = Key
        self.user = User

        self.set_account(Account)

        c = self.get_cache(self.profile['name'])
        if c == None:
            self.valid = False
            self.cached = None
        else:
            self.valid = True
            self.cached = c['instances']

    def do_fzf(self, instances):
        sel = iterfzf(instances)

        if sel is None:
            return

        name, ip, account, tags = sel.split('\t')

        ssh_command = 'ssh '
        if self.key:
            ssh_command += f" -i {self.key} "
        if self.user:
            ssh_command += f" -l {self.user} "

        ssh_command += f" {ip}"

        self.ctx.vlog("Executing:")
        self.ctx.log(ssh_command)

        self.do_ssh(ssh_command)

    def do_ssh(self, command):
        subprocess.call(
            command, shell=True, executable=os.getenv('SHELL', '/bin/bash'))


def to_fzf_format(l):
    out = []
    for i in l:
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


@click.command("ssh")
@common.common_params(p=False)
@click.option(
    '-u',
    '--user',
    default=None,
    show_default="''",
    show_envvar=True,
    help='Username to use with SSH')
@click.option(
    '-k',
    '--key',
    default=None,
    show_default="''",
    show_envvar=True,
    help='SSH key path')
@common.cache_params()
@pass_environment
def cli(ctx, **kwargs):
    """SSH to EC2 instance"""

    ssh = SSH(ctx, kwargs['cache'], kwargs['account'], kwargs['key'],
              kwargs['user'], kwargs['cache_time'])

    if ssh.valid:
        # Found cached results
        ret = ssh.results
    else:
        # Scan account
        _service = "AWS::EC2::Instance"
        _select = "resourceId, accountId, configuration.privateIpAddress, tags"

        _filter = f"resourceType like '{_service}' AND " \
             "configuration.state.name like 'running'"

        if kwargs['filter']:
            _filter += f" AND {kwargs['filter']}"

        if kwargs['account'] != 'all':
            _filter += f" AND accountId like '{ssh.account_id}'"

        if kwargs['region'] != 'all':
            _filter += f" AND awsRegion like '{kwargs['region']}'"

        query = Query(
            ctx,
            Service=_service,
            Select=_select,
            Filter=_filter,
            Limit=kwargs['limit'],
            Account=kwargs['account'],
            Region=kwargs['region'],
            Pager=kwargs['pager'])

        if query.valid:
            # Found cached results
            if kwargs['pager']:
                ctx.log(query.cached)
                return
            else:
                tmp = query.cached
        else:
            # Query instances
            tmp = query.query(kwargs['cache_time'])

        ssh.do_fzf(to_fzf_format(tmp))


def do_query(ctx, kwargs):
    ctx.vlog(kwargs)
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
