import click
import functools


def common_params(a="all", r="all", f="''", p=True, l=0):
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
            '--pager',
            'pager',
            flag_value=True,
            default=p,
            show_default=True,
            help='Send query results to pager')
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
