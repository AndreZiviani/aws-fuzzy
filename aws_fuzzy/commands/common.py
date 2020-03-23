import click
import functools


def common_params(func):
    @click.option(
        '-a',
        '--account',
        default='all',
        help='Filter by accountid, defaults to every account')
    @click.option(
        '-r',
        '--region',
        default='all',
        help='Filter by region, defaults to every region')
    @click.option(
        '-f',
        '--filter',
        default='',
        help='Use a custom query to filter results')
    @click.option(
        '--pager',
        'pager',
        flag_value=True,
        default=True,
        help='Send query results to pager')
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)

    return wrapper
