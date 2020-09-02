import functools
import re
import shelve
import json
import glob
import os

from subprocess import run
from datetime import datetime
from os.path import expanduser

import click
import boto3


def p_account(account="all",
              show_default=True,
              show_envvar=True,
              msg='Filter by accountid'):
    def params(func):
        @click.option(
            '-a',
            '--account',
            default=account,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_select(select=None,
             show_default=True,
             show_envvar=True,
             msg='Custom select to filter results'):
    def params(func):
        @click.option(
            '-s',
            '--select',
            default=select,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_region(region='all',
             show_default=True,
             show_envvar=True,
             msg='Filter by region'):
    def params(func):
        @click.option(
            '-r',
            '--region',
            default=region,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_filter(filter='',
             show_default=True,
             show_envvar=True,
             msg='Use a custom query to filter results'):
    def params(func):
        @click.option(
            '-f',
            '--filter',
            default=filter,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_pager(pager=True,
            show_default=True,
            show_envvar=True,
            msg='Send query results to pager'):
    def params(func):
        @click.option(
            '--pager/--no-pager',
            'pager',
            flag_value=True,
            default=pager,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_limit(limit=0,
            show_default=True,
            show_envvar=True,
            msg='Limit the number of results'):
    def params(func):
        @click.option(
            '-l',
            '--limit',
            default=limit,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_cache(cache=True,
            show_default=True,
            show_envvar=True,
            msg='Whether to use cached results'):
    def params(func):
        @click.option(
            '--cache/--no-cache',
            'cache',
            flag_value=True,
            default=cache,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_cache_time(cache_time=3600,
                 show_default=True,
                 show_envvar=True,
                 msg='Cache results TTL in seconds'):
    def params(func):
        @click.option(
            '--cache-time',
            default=cache_time,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_inventory(inventory=None,
                show_default='First one found',
                show_envvar=True,
                msg='What inventory to use'):
    def params(func):
        @click.option(
            '-i',
            '--inventory',
            default=inventory,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def p_profile(
        profile=os.getenv('AWS_PROFILE', 'default'),
        show_default='$AWS_PROFILE',
        show_envvar=True,
        msg='What profile to use'):
    def params(func):
        @click.option(
            '-p',
            '--profile',
            default=profile,
            show_default=show_default,
            show_envvar=show_envvar,
            help=msg)
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


class Common():
    def __init__(self, ctx):
        self.aws_dir = expanduser("~") + "/.aws"
        self.sso_dir = self.aws_dir + "/sso/cache"
        self.profiles, self.account_ids = self._get_profiles()

        self.profile = None
        self.account_id = None

        self.ctx = ctx

    def check_expired(self, date):
        now = datetime.utcnow()
        return date < now

    def _get_profiles(self):
        d = {}
        rd = {}
        with open(f"{self.aws_dir}/config", 'r') as f:
            for l in f:
                if not l.strip():  #if empty line
                    continue

                if l[0] == '[':
                    tmp = re.findall(r'\[(.*)\]', l)[0]
                    if 'profile' in tmp:
                        profile_name = str(tmp.split()[1])
                    else:
                        profile_name = str(tmp)
                    d[profile_name] = {"name": profile_name}
                    continue
                key, value = re.findall(r'([a-zA-Z_]+)\s*=\s*(.*)', l)[0]
                d[profile_name][key] = value
        for k, v in d.items():
            account_id = d[k]['sso_account_id']
            rd[account_id] = v
        return d, rd

    def set_account(self, account):
        if account == 'all':
            self.profile = 'all'
            self.account_id = 00000000000
        else:
            if re.match("[0-9]+", str(account)):
                # Got account ID
                self.account_id = account
                for k in self.profiles:
                    if self.profiles[k]['sso_account_id'] == account:
                        self.profile = self.profiles[k]
            else:
                # Got account name
                p = self.profiles[account]
                self.profile = p
                self.account_id = p['sso_account_id']


class Cache(Common):
    def __init__(self, ctx, Service, Cache_time=3600):
        super().__init__(ctx)
        self.cache_dir = expanduser("~") + "/.aws-fuzzy"

        self.service = Service
        self.cache = False
        self.cache_time = Cache_time
        self.remove_expired_items()

    def remove_expired_items(self):
        if self.cache:
            count = 1
            s = shelve.open(f"{self.cache_dir}/{self.service}")
            for item in s:
                if self.check_expired(s[item]["expires"]):
                    count += 1
                    del s[item]
            self.ctx.vlog(f"Removed {count} expired items from cache")
            s.close()

    def get_cache(self, item):
        if self.cache:
            s = shelve.open(f"{self.cache_dir}/{self.service}")
            if item in s:
                if self.check_expired(s[item]["expires"]):
                    del s[item]
                else:
                    tmp = s[item]
                    s.close()
                    self.ctx.vlog("Using cached results!")
                    return tmp
            else:
                self.ctx.vlog("Missing from cache")

            s.close()
        return None

    def set_cache(self, item, value):
        if self.cache:
            s = shelve.open(f"{self.cache_dir}/{self.service}")
            s[item] = value
            s.close()


class SSO(Cache):
    def __init__(self, ctx, Cache, Account, Cache_time):
        super().__init__(ctx, "sso", Cache_time)

        self.cache = Cache

        self.set_account(Account)

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

        self.sso_token = self.get_sso_token()

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
