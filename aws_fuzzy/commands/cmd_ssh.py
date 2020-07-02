from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params
from .common import cache_params
from .common import get_profile
from .common import get_cache
from .common import set_cache

import click
import os
import subprocess
import boto3
import re

from iterfzf import iterfzf
from datetime import datetime
from datetime import timedelta


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
    default="''",
    show_default=True,
    show_envvar=True,
    help='Username to use with SSH')
@click.option(
    '-k',
    '--key',
    default="''",
    show_default=True,
    show_envvar=True,
    help='SSH key path')
@cache_params()
@pass_environment
def cli(ctx, **kwargs):
    """SSH to EC2 instance"""

    profile = get_profile(kwargs['profile'])['name']

    if kwargs['cache']:
        ret = get_cache(ctx, "ssh", profile)
        if ret != None:
            instances = ret['instances']
        else:
            instances = get_instances()
            tmp = {
                'instances':
                instances,
                'expires':
                datetime.utcnow() + timedelta(seconds=kwargs['cache_time'])
            }
            set_cache(ctx, "ssh", profile, tmp)

    else:
        instances = get_instances()

    do_fzf(ctx, kwargs, instances)


def get_instances():
    ec2 = boto3.client('ec2')
    all_instances = ec2.describe_instances(
        Filters=[{
            'Name': 'instance-state-name',
            'Values': ['running']
        }])['Reservations']

    instances = []

    for tmp in all_instances:
        for i in tmp['Instances']:
            name = get_tag_value(i['Tags'], 'Name')
            ip = i['NetworkInterfaces'][0]['PrivateIpAddress']
            instance_id = i['InstanceId']
            instances.append(f"{name} ({instance_id}) @ {ip}")

    return instances


def get_tag_value(tags, key):
    for t in tags:
        if key in t['Key']:
            return t['Value'].replace('"', '')

    return "Unnamed Instance"


def do_fzf(ctx, kwargs, instances):
    sel = iterfzf(instances)

    if sel is None:
        return

    ctx.vlog(sel)
    name, ip = re.findall('([\w-]+) \(i-\w+\) @ (.*)', sel)[0]

    if kwargs['key'] != "''":
        key = f"-i {kwargs['key']}"
    else:
        key = ''

    if kwargs['user'] != "''":
        user = f"-l {kwargs['user']}"
    else:
        user = ''

    ssh_command = f"ssh {key} {user} {ip} # {name}"

    ctx.vlog("Executing:")
    ctx.vlog(ssh_command)

    subprocess.call(
        ssh_command, shell=True, executable=os.getenv('SHELL', '/bin/bash'))
