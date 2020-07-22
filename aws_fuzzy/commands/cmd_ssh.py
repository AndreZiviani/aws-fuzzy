import os
import subprocess
import re

from datetime import datetime
from datetime import timedelta
from iterfzf import iterfzf

import click
import boto3

from aws_fuzzy.cli import pass_environment
from aws_fuzzy import common


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

        self.instances = None

        c = self.get_cache(self.profile['name'])
        if c is None:
            self.list_instances()
        else:
            self.instances = c['instances']

    def get_tag_value(self, tags, key):
        for t in tags:
            if key in t['Key']:
                return t['Value'].replace('"', '')

        return "Unnamed Instance"

    def list_instances(self):
        ec2 = boto3.client('ec2')
        all_instances = ec2.describe_instances(
            Filters=[{
                'Name': 'instance-state-name',
                'Values': ['running']
            }])['Reservations']

        instances = []

        for tmp in all_instances:
            for i in tmp['Instances']:
                name = self.get_tag_value(i['Tags'], 'Name')
                ip = i['PrivateIpAddress']
                instance_id = i['InstanceId']
                instances.append(f"{name} ({instance_id}) @ {ip}")

        self.instances = instances
        expires = datetime.utcnow() + timedelta(seconds=self.cache_time)
        self.set_cache(self.profile['name'], {
            'instances': instances,
            'expires': expires
        })

        return instances

    def do_fzf(self, instances):
        sel = iterfzf(instances)

        if sel is None:
            return

        name, ip = re.findall(r'([\w-]+) \(i-\w+\) @ (.*)', sel)[0]

        ssh_command = 'ssh '
        if self.key:
            ssh_command += f" -i {self.key} "
        if self.user:
            ssh_command += f" -l {self.user} "

        ssh_command += f" {ip} # {name}"

        self.ctx.vlog("Executing:")
        self.ctx.log(ssh_command)

        self.do_ssh(ssh_command)

    def do_ssh(self, command):
        subprocess.call(
            command, shell=True, executable=os.getenv('SHELL', '/bin/bash'))


@click.command("ssh")
@click.option(
    '-p',
    '--profile',
    default=os.getenv('AWS_PROFILE', 'default'),
    show_default="$AWS_PROFILE",
    show_envvar=True,
    help='AWS Profile')
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
@common.p_cache()
@common.p_cache_time()
@pass_environment
def cli(ctx, **kwargs):
    """SSH to EC2 instance"""

    ssh = SSH(ctx, kwargs['cache'], kwargs['profile'], kwargs['key'],
              kwargs['user'], kwargs['cache_time'])

    ssh.do_fzf(ssh.instances)
