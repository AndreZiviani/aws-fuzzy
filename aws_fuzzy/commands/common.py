import click
import functools
import re
import shelve

from datetime import datetime
from os.path import expanduser

ENVVAR_PREFIX = "AWSFUZZY"

AWS_DIR = expanduser("~") + "/.aws"
SSO_CRED_DIR = AWS_DIR + "/sso/cache"
SSO_PROFILES = AWS_DIR + "/config"


def common_params(a="all", s=None, r="all", f="''", p=True, c=True, l=0):
    def params(func):
        @click.option(
            '-a',
            '--account',
            default=a,
            show_default=True,
            show_envvar=True,
            help='Filter by accountid')
        @click.option(
            '-s',
            '--select',
            default=s,
            show_default=True,
            show_envvar=True,
            help='Custom select to filter results')
        @click.option(
            '-r',
            '--region',
            default=r,
            show_default=True,
            show_envvar=True,
            help='Filter by region')
        @click.option(
            '-f',
            '--filter',
            default=f,
            show_default=True,
            show_envvar=True,
            help='Use a custom query to filter results')
        @click.option(
            '--pager/--no-pager',
            'pager',
            flag_value=True,
            default=p,
            show_default=True,
            show_envvar=True,
            help='Send query results to pager')
        @click.option(
            '-l',
            '--limit',
            default=l,
            show_default=True,
            show_envvar=True,
            help='Use a custom query to filter results')
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def cache_params(cache=True, cache_time=3600):
    def params(func):
        @click.option(
            '--cache/--no-cache',
            'cache',
            flag_value=True,
            default=cache,
            show_default=True,
            show_envvar=True,
            help='Whether to use cached results')
        @click.option(
            '--cache-time',
            default=cache_time,
            show_default=True,
            show_envvar=True,
            help='Cache results TTL in seconds')
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def modify_none(ctx, param, value):
    if value == "_none_":
        return None
    return value


def query_params():
    def params(func):
        @click.option(
            '-i',
            '--inventory',
            default="_none_",
            show_default="First one found",
            show_envvar=True,
            callback=modify_none,
            help='Cache results TTL in seconds')
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params


def get_profile(profile):
    path = SSO_PROFILES
    d = {}
    with open(path, 'r') as f:
        for l in f:
            if l[0] == '[':
                profile_name = ''
                tmp = re.findall('\[(.*)\]', l)[0]
                if 'profile' in tmp:
                    profile_name = tmp.split()[1]
                elif tmp == 'default':
                    profile_name = 'default'
                d[profile_name] = {}
                d[profile_name]["name"] = profile_name
                continue
            else:
                if l.strip() == '':
                    continue
                key, value = re.findall('([a-zA-Z_]+)\s*=\s*(.*)', l)[0]
                d[profile_name][key] = value

    try:
        if re.match("[0-9]+", str(profile)):
            # Got account ID
            for k in d:
                if d[k]['sso_account_id'] == profile:
                    return d[k]
        else:
            # Got account name
            return d[profile]

    except KeyError:
        return None


def check_expired(date):
    now = datetime.utcnow()
    if date < now:
        return True
    else:
        return False


def get_cache(ctx, service, item):
    s = shelve.open(ctx.cache_dir + f"/{service}")
    if item in s:
        if check_expired(s[item]["expires"]):
            del s[item]
        else:
            tmp = s[item]
            s.close()
            ctx.vlog("Using cached results!")
            return tmp

    s.close()
    return None


def set_cache(ctx, service, item, value):
    s = shelve.open(ctx.cache_dir + f"/{service}")
    s[item] = value
    s.close()
