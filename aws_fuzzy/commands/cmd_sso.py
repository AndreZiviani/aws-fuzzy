from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params
from .common import cache_params
from .common import get_profile
from .common import check_expired

import click
import re
import os
import re
import glob
import datetime
import json
import shelve

import boto3
from iterfzf import iterfzf
from os.path import expanduser
from subprocess import run
from datetime import datetime
from datetime import timedelta

AWS_DIR = expanduser("~") + "/.aws"
SSO_CRED_DIR = AWS_DIR + "/sso/cache"
SSO_PROFILES = AWS_DIR + "/config"


def get_latest_credential(path):
    list_of_files = glob.glob(path + '/*')
    return max(list_of_files, key=os.path.getctime)


def get_sso_credentials(path):
    latest_cred = get_latest_credential(path)

    with open(latest_cred, 'r') as f:
        raw = f.read()

    j = json.loads(raw)

    expires = datetime.strptime(j['expiresAt'], '%Y-%m-%dT%H:%M:%SUTC')
    if check_expired(expires, 0):
        raise KeyError

    return j["accessToken"]


def print_credentials(creds):
    expires = datetime.utcfromtimestamp(int(
        creds['expiration'] / 1000)).strftime('%Y-%m-%dT%H:%M:%SUTC')
    print(f"""
export AWS_ACCESS_KEY_ID='{creds['accessKeyId']}';
export AWS_SECRET_ACCESS_KEY='{creds['secretAccessKey']}';
export AWS_SESSION_TOKEN='{creds['sessionToken']}';
export AWS_SECURITY_TOKEN='{creds['sessionToken']}';
export AWS_EXPIRES='{expires}';
    """)


@click.group("sso")
@click.pass_context
def cli(ctx, **kwargs):
    """AWS SSO service"""


@cli.command()
@click.option(
    '-p',
    '--profile',
    default=os.getenv('AWS_PROFILE', 'default'),
    show_default="$AWS_PROFILE",
    help='AWS Profile')
@cache_params()
@pass_environment
def login(ctx, **kwargs):
    """Login to AWS SSO"""

    p = get_profile(kwargs['profile'])

    ctx.vlog(kwargs)
    ctx.vlog(p)
    if kwargs['cache']:
        s = shelve.open(ctx.cache_dir + "/sso")
        if p['name'] in s:
            ctx.vlog('Found profile in cache')
            c = s[p['name']]
            if check_expired(c['expires'], 0):
                ctx.log(
                    'Found credentials in cache but they are expired, requesting new one'
                )
            else:
                print_credentials(c['credentials'])
                return

    try:
        sso_token = get_sso_credentials(SSO_CRED_DIR)
    except KeyError:
        ctx.log("Failed to get SSO credentials, trying to authenticate again")
        ret = run(['aws', 'sso', 'login'])
        if ret.returncode != 0:
            ctx.log("Something went wrong trying to login")
            return
        sso_token = get_sso_credentials(SSO_CRED_DIR)

    sso = boto3.client('sso')

    try:
        ret = sso.get_role_credentials(
            roleName=p["sso_role_name"],
            accountId=p["sso_account_id"],
            accessToken=sso_token)
        credentials = ret["roleCredentials"]
        if kwargs['cache']:
            expires = datetime.utcfromtimestamp(
                int(credentials['expiration'] / 1000))
            s[p['name']] = {'credentials': credentials, 'expires': expires}
        print_credentials(credentials)
    except Exception as e:
        ctx.log("Invalid SSO token, removing credentials")
