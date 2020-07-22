import os
import glob
import json

from subprocess import run
from datetime import datetime
from os.path import expanduser

import click
import boto3

from aws_fuzzy.cli import pass_environment
from aws_fuzzy import common


@click.group("sso")
@click.pass_context
def cli(ctx, **kwargs):
    """AWS SSO service"""


class SSO(common.Cache):
    def __init__(self, ctx, Cache, Account, Cache_time):
        super().__init__(ctx, "sso", Cache_time)

        self.cache = Cache

        self.set_account(Account)

        self.sso_token = self.get_sso_token()

        c = self.get_cache(self.profile['name'])
        if c is None:
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

            self.sso_token = j["accessToken"]
            return j["accessToken"]
        except:
            self.ctx.log(
                "Failed to get SSO credentials, trying to authenticate again")
            ret = run(['aws', 'sso', 'login'],
                      stdout=click.get_text_stream('stderr'),
                      check=True)
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

        client = boto3.client('sso')
        ret = client.get_role_credentials(
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

    def list_accounts(self, maxResults=100, region='us-east-1', profile=None):
        if profile is None:
            profile = self.profile['name']

        session = boto3.Session(profile_name=profile)
        client = session.client('sso', region_name=region)

        ret = client.list_accounts(
            accessToken=self.sso_token, maxResults=maxResults)['accountList']

        accounts = []
        for account in ret:
            d = {}
            d['name'] = account['accountName'].replace(' ', '_')
            d['id'] = account['accountId']
            roles = client.list_account_roles(
                accountId=account['accountId'],
                accessToken=self.sso_token,
                maxResults=maxResults)['roleList']

            if len(roles) > 1:
                r = []
                for role in roles:
                    r.append(role['roleName'])

                role = click.prompt(
                    f"Found multiple roles for account '{account['accountName']} ({account['accountId']})': {r}"
                )
            else:
                role = roles[0]['roleName']

            d['role'] = role
            accounts.append(d)

        return accounts


@cli.command()
@click.option(
    '-p',
    '--profile',
    default=os.getenv('AWS_PROFILE', 'default'),
    show_default="$AWS_PROFILE",
    show_envvar=True,
    help='AWS Profile')
@common.p_cache()
@common.p_cache_time()
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

    # Missing or expired cached credentials, requesting new one
    ctx.vlog("Could not find cached credentials or they are expired")
    sso.get_new_credentials()
    sso.print_credentials()


@cli.command()
@common.p_region(region='us-east-1')
@common.p_cache()
@common.p_cache_time()
@pass_environment
def configure(ctx, **kwargs):
    """Configure AWS SSO"""

    sso_url = None
    sso_profiles = expanduser('~') + '/.aws/config'
    sso_region = kwargs['region']

    if os.path.isfile(sso_profiles):
        ctx.log(
            "AWS config file already exists, making a backup before updating")
        p = common.Common(ctx)

        if 'default' in p.profiles:
            sso_url = p.profiles['default']['sso_start_url']
            sso_region = p.profiles['default']['sso_region']

        os.rename(sso_profiles, f"{sso_profiles}.bkp")

    start = click.prompt("Enter SSO start url", default=sso_url)
    region = click.prompt("Default region", default=sso_region)

    f = open(sso_profiles, 'w')
    f.write(f"""[default]
sso_start_url = {start}
sso_region = {region}
sso_role_name = dummy
sso_account_id = 00000000""")
    f.close()

    sso = SSO(
        ctx,
        Cache=kwargs['cache'],
        Account='default',
        Cache_time=kwargs['cache_time'])

    accounts = sso.list_accounts(region=sso_region, profile='default')

    f = open(sso_profiles, 'a')
    for account in accounts:
        f.write(f"""
[profile {account['name']}]
sso_start_url = {start}
sso_region = {region}
sso_account_id = {account['id']}
sso_role_name = {account['role']}""")

    f.close()

    ctx.log("Done! Please check ~/.aws/config for your profiles")
