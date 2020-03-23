from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params

import click
import re
import os
import re
import glob
import datetime
import json

import boto3
from iterfzf import iterfzf
from os.path import expanduser

AWS_DIR = expanduser("~") + "/.aws"
SSO_CRED_DIR = AWS_DIR + "/sso/cache"
SSO_PROFILES = AWS_DIR + "/config"


def get_profile(profile, path):
    d = {}
    with open(path, 'r') as f:
        for l in f:
            if l[0] == '[':
                profile = re.findall('\[(.*)\]', l)[0]
                if 'profile' in profile:
                    profile = profile.split()[1]
                d[profile] = {}
                d[profile]["name"] = profile
                continue
            else:
                key, value = re.findall('([a-zA-Z_]+)\s*=\s*(.*)', l)[0]
                d[profile][key] = value

    return d[profile]


def check_expired(cred):
    now = datetime.datetime.utcnow()
    expires = datetime.datetime.strptime(cred, '%Y-%m-%dT%H:%M:%SUTC')
    if expires < now:
        return True
    else:
        return False


def get_latest_credential(path):
    list_of_files = glob.glob(path + '/*')
    return max(list_of_files, key=os.path.getctime)


def get_sso_credentials(path):
    latest_cred = get_latest_credential(path)

    with open(latest_cred, 'r') as f:
        raw = f.read()

    j = json.loads(raw)

    if check_expired(j["expiresAt"]):
        print('x')
        raise KeyError

    return j["accessToken"]


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
@pass_environment
def login(ctx, **kwargs):
    """Login to AWS SSO"""

    p = get_profile(kwargs['profile'], SSO_PROFILES)

    try:
        sso_token = get_sso_credentials(SSO_CRED_DIR)
    except KeyError:
        ctx.log("Failed to get SSO credentials, trying to authenticate again")

    try:
        sso = boto3.client('sso')
        ret = sso.get_role_credentials(
            roleName=p["sso_role_name"],
            accountId=p["sso_account_id"],
            accessToken=sso_token)
        credentials = ret["roleCredentials"]
        print(credentials)
    except Exception as e:
        ctx.log(e)
        ctx.log("Invalid SSO token, removing credentials")
        #os.remove(get_latest_credential(SSO_CRED_DIR))
