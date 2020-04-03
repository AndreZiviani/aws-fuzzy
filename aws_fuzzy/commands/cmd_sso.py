from aws_fuzzy.cli import pass_environment
from .common import cache_params
from .common import get_profile
from .common import check_expired
from .common import get_cache
from .common import set_cache

import click
import os
import glob
import json

import boto3
from subprocess import run
from datetime import datetime
from os.path import expanduser

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
    show_envvar=True,
    help='AWS Profile')
@cache_params()
@pass_environment
def login(ctx, **kwargs):
    """Login to AWS SSO"""

    p = get_profile(kwargs['profile'])

    ctx.vlog(kwargs)
    ctx.vlog(p)
    if kwargs['cache']:
        ret = get_cache(ctx, "sso", p['name'])

        if ret != None:
            print_credentials(ret['credentials'])
            return

        ctx.vlog("Could not find cached credentials or they are expired")

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
            set_cache(ctx, "sso", p['name'], {
                'credentials': credentials,
                'expires': expires
            })
        print_credentials(credentials)
    except Exception:
        ctx.log("Invalid SSO token, removing credentials")
