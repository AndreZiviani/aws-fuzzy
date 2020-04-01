import click
import functools


AWS_DIR = expanduser("~") + "/.aws"
SSO_CRED_DIR = AWS_DIR + "/sso/cache"
SSO_PROFILES = AWS_DIR + "/config"


def common_params(a="all", r="all", f="''", p=True, c=True, l=0):
    def params(func):
        @click.option(
            '-a',
            '--account',
            default=a,
            show_default=True,
            help='Filter by accountid')
        @click.option(
            '-r',
            '--region',
            default=r,
            show_default=True,
            help='Filter by region')
        @click.option(
            '-f',
            '--filter',
            default=f,
            show_default=True,
            help='Use a custom query to filter results')
        @click.option(
            '--pager/--no-pager',
            'pager',
            flag_value=True,
            default=p,
            show_default=True,
            help='Send query results to pager')
        @click.option(
            '--cache/--no-cache',
            'cache',
            flag_value=True,
            default=c,
            show_default=True,
            help='Whether to use cached results')
        @click.option(
            '-l',
            '--limit',
            default=l,
            show_default=True,
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
            help='Whether to use cached results')
        @click.option(
            '--cache-time',
            default=cache_time,
            show_default=True,
            help='Cache results TTL in seconds')
        @functools.wraps(func)
        def wrapper(*args, **kwargs):
            return func(*args, **kwargs)

        return wrapper

    return params
