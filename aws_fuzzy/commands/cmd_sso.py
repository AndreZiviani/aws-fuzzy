from aws_fuzzy.cli import pass_environment
from aws_fuzzy import common

import click
import os
import glob
import json

import boto3
from subprocess import run
from datetime import datetime
from os.path import expanduser


@click.group("sso")
@click.pass_context
def cli(ctx, **kwargs):
    """AWS SSO service"""


class SSO(common.Cache):
    def __init__(self, ctx, Cache, Account, Cache_time):
        super().__init__(ctx, "sso", Cache_time)

        self.cache = Cache
        self.client = boto3.client('sso')

        self.set_account(Account)

        self.sso_token = self.get_sso_token()

        c = self.get_cache(self.profile['name'])
        if c == None:
            self.valid = False
            self.access_key = None
            self.secret_key = None
            self.session_token = None
            self.expiration = None
        else:
            c = c['credentials']
            self.valid = True
            self.access_key = c['accessKeyId']
            self.secret_key = c['secretAccessKey']
            self.session_token = c['sessionToken']
            self.expiration = c['expiration']

    def get_sso_token(self):
        list_of_files = glob.glob(f"{self.sso_dir}/*")
        try:
            latest_cred = max(list_of_files, key=os.path.getctime)

            with open(latest_cred, 'r') as f:
                raw = f.read()

            j = json.loads(raw)

            expires = datetime.strptime(j['expiresAt'], '%Y-%m-%dT%H:%M:%SUTC')
            if self.check_expired(expires):
                raise
            else:
                self.sso_token = j["accessToken"]
                return j["accessToken"]
        except:
            self.ctx.log(
                "Failed to get SSO credentials, trying to authenticate again")
            ret = run(['aws', 'sso', 'login'],
                      stdout=click.get_text_stream('stderr'))
            if ret.returncode != 0:
                self.ctx.log("Something went wrong trying to login")
                return None
            self.sso_token = self.get_sso_token()

            return self.sso_token

    def set_credentials(self, credentials, expires):
        self.set_cache(self.profile['name'], {
            "credentials": credentials,
            "expires": expires,
        })
        self.valid = True
        self.access_key = credentials['accessKeyId']
        self.secret_key = credentials['secretAccessKey']
        self.session_token = credentials['sessionToken']
        self.expiration = credentials['expiration']

    def get_new_credentials(self):

        ret = self.client.get_role_credentials(
            roleName=self.profile["sso_role_name"],
            accountId=self.profile["sso_account_id"],
            accessToken=self.sso_token)

        credentials = ret["roleCredentials"]

        expires = datetime.utcfromtimestamp(
            int(credentials['expiration'] / 1000))

        self.set_credentials(credentials, expires)

        return credentials

    def print_credentials(self):
        expires = datetime.utcfromtimestamp(int(
            self.expiration / 1000)).strftime('%Y-%m-%dT%H:%M:%SUTC')
        print(f"""
    export AWS_ACCESS_KEY_ID='{self.access_key}';
    export AWS_SECRET_ACCESS_KEY='{self.secret_key}';
    export AWS_SESSION_TOKEN='{self.session_token}';
    export AWS_SECURITY_TOKEN='{self.session_token}';
    export AWS_EXPIRES='{expires}';
        """)


@cli.command()
@click.option(
    '-p',
    '--profile',
    default=os.getenv('AWS_PROFILE', 'default'),
    show_default="$AWS_PROFILE",
    show_envvar=True,
    help='AWS Profile')
@common.cache_params()
@pass_environment
def login(ctx, **kwargs):
    """Login to AWS SSO"""

    sso = SSO(
        ctx,
        Cache=kwargs['cache'],
        Account=kwargs['profile'],
        Cache_time=kwargs['cache_time'])

    ctx.vlog(kwargs)
    if sso.valid:
        # We got valid cached credentials
        sso.print_credentials()
        return
    else:
        # Missing or expired cached credentials, requesting new one
        ctx.vlog("Could not find cached credentials or they are expired")
        sso.get_new_credentials()
        sso.print_credentials()
