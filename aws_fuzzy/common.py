import click
import functools
import re
import shelve

from datetime import datetime
from os.path import expanduser

from aws_fuzzy.cli import pass_environment


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


class Common():
    def __init__(self, ctx):
        self.aws_dir = expanduser("~") + "/.aws"
        self.sso_dir = self.aws_dir + "/sso/cache"
        self.profiles = self._get_profiles()

        self.profile = None
        self.account_id = None

        self.ctx = ctx

    def check_expired(self, date):
        now = datetime.utcnow()
        if date < now:
            return True
        else:
            return False

    def _get_profiles(self):
        d = {}
        with open(f"{self.aws_dir}/config", 'r') as f:
            for l in f:
                if not l.strip():  #if empty line
                    continue

                if l[0] == '[':
                    tmp = re.findall('\[(.*)\]', l)[0]
                    if 'profile' in tmp:
                        profile_name = str(tmp.split()[1])
                    else:
                        profile_name = str(tmp)
                    d[profile_name] = {"name": profile_name}
                    continue
                else:
                    key, value = re.findall('([a-zA-Z_]+)\s*=\s*(.*)', l)[0]
                    d[profile_name][key] = value
        return d

    def set_account(self, account):
        if account == 'all':
            self.profile = 'all'
            self.account_id = 00000000000
        else:
            if re.match("[0-9]+", str(account)):
                # Got account ID
                self.account_id = account
                for k in self.profiles:
                    if self.profiles[k]['sso_account_id'] == profile:
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
        if self.cache == True:
            count = 1
            s = shelve.open(f"{self.cache_dir}/{self.service}")
            for item in s:
                if self.check_expired(s[item]["expires"]):
                    count += 1
                    del s[i]
            self.ctx.vlog(f"Removed {count} expired items from cache")
            s.close()

    def get_cache(self, item):
        if self.cache == True:
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
        if self.cache == True:
            s = shelve.open(f"{self.cache_dir}/{self.service}")
            s[item] = value
            s.close()
