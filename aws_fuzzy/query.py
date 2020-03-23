import boto3
import json
import click

from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.lexers import PythonLexer
from pygments.formatters import TerminalFormatter


def do_query(Expression=None, ConfigurationAggregatorName=None, Limit=None):
    c = boto3.client('config')

    if Limit <= 100:
        o = c.select_aggregate_resource_config(
            Expression=Expression,
            ConfigurationAggregatorName=ConfigurationAggregatorName,
            Limit=Limit)
    else:  # Iterate through pages until Limit is reached or end of results
        it = Limit / 100
        mod = Limit % 100

        o = c.select_aggregate_resource_config(
            Expression=Expression,
            ConfigurationAggregatorName=ConfigurationAggregatorName,
            Limit=100)

        tmp = o
        i = 1
        while 'NextToken' in tmp and i < it:
            tmp = c.select_aggregate_resource_config(
                Expression=Expression,
                ConfigurationAggregatorName=ConfigurationAggregatorName,
                Limit=100,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])

        if mod > 0:
            tmp = c.select_aggregate_resource_config(
                Expression=Expression,
                ConfigurationAggregatorName=ConfigurationAggregatorName,
                Limit=mod,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])

    j = [json.loads(r) for r in o['Results']]

    return j


def query(ctx, **kwargs):

    if 'select' not in kwargs:
        kwargs['select'] = "resourceId, accountId, configuration, tags"

    if kwargs['filter'] != "''":
        kwargs[
            'filter'] = f"resourceType like '{kwargs['service']}' AND {kwargs['filter']}"
    else:
        kwargs['filter'] = f"resourceType like '{kwargs['service']}'"

    kwargs[
        'expression'] = f"SELECT {kwargs['select']} WHERE {kwargs['filter']}"

    ctx.vlog("kwargs:")
    ctx.vlog(kwargs)

    ret = do_query(kwargs['expression'], 'linx-digital-inventory-assets',
                   kwargs['limit'])

    if kwargs['pager']:
        click.echo_via_pager(
            highlight(
                json.dumps(ret, indent=4), JsonLexer(), TerminalFormatter()))

    return ret
