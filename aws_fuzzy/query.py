import boto3
import os
import json
import click
import sys

from .commands.common import get_profile
from .commands.common import get_cache
from .commands.common import set_cache

from datetime import timedelta
from datetime import datetime
from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.formatters import TerminalFormatter


def do_query(ctx,
             cache_time,
             Expression=None,
             ConfigurationAggregatorName=None,
             Limit=None):
    c = boto3.client('config')

    if ConfigurationAggregatorName == None:
        ret = get_cache(ctx, "inventory", os.getenv('AWS_PROFILE', "unknown"))
        if ret != None:
            ConfigurationAggregatorName = ret['inventory']

        else:
            aggs = c.describe_configuration_aggregators()
            ConfigurationAggregatorName = aggs['ConfigurationAggregators'][0][
                'ConfigurationAggregatorName']

            tmp = {
                'inventory': ConfigurationAggregatorName,
                'expires': datetime.utcnow() + timedelta(seconds=cache_time)
            }
            set_cache(ctx, "inventory", os.getenv('AWS_PROFILE', "unknown"),
                      tmp)

    if Limit <= 100 and Limit > 0:
        o = c.select_aggregate_resource_config(
            Expression=Expression,
            ConfigurationAggregatorName=ConfigurationAggregatorName,
            Limit=Limit)
        t = len(o['Results'])
    else:  # Iterate through pages until Limit is reached or end of results
        if Limit > 0:
            it = Limit / 100
            mod = Limit % 100
        else:
            it = sys.maxsize
            mod = 0

        o = c.select_aggregate_resource_config(
            Expression=Expression,
            ConfigurationAggregatorName=ConfigurationAggregatorName,
            Limit=100)

        tmp = o
        i = 1
        r = len(o['Results'])
        t = r
        ctx.vlog(f'Got {r} results')
        while 'NextToken' in tmp and i < it:
            tmp = c.select_aggregate_resource_config(
                Expression=Expression,
                ConfigurationAggregatorName=ConfigurationAggregatorName,
                Limit=100,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])
            i += 1
            r = len(tmp['Results'])
            t += r
            ctx.vlog(f'Got {r} results')

        if mod > 0:
            tmp = c.select_aggregate_resource_config(
                Expression=Expression,
                ConfigurationAggregatorName=ConfigurationAggregatorName,
                Limit=mod,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])
            r = len(tmp['Results'])
            t += r
            ctx.vlog(f'Got {r} results')

    ctx.vlog(f'Got a total of {t} results')

    j = [json.loads(r) for r in o['Results']]

    return j


def query(ctx, kwargs):

    params = kwargs
    if kwargs['select'] == None:
        params[
            'select'] = "resourceId, accountId, awsRegion, configuration, tags"

    if kwargs['filter'] != "''":
        params[
            'filter'] = f"resourceType like '{kwargs['service']}' AND {kwargs['filter']}"
        if kwargs['account'] != 'all':
            account = get_profile(kwargs['account'])
            params[
                'filter'] += f" AND accountId like '{account['sso_account_id']}'"
    else:
        params['filter'] = f"resourceType like '{kwargs['service']}'"
        if kwargs['account'] != 'all':
            account = get_profile(kwargs['account'])
            params[
                'filter'] += f" AND accountId like '{account['sso_account_id']}'"
        if kwargs['region'] != 'all':
            params['filter'] += f" AND awsRegion like '{kwargs['account']}'"

    params[
        'expression'] = f"SELECT {kwargs['select']} WHERE {kwargs['filter']}"

    ctx.vlog("params:")
    ctx.vlog(params)

    ret = get_cache(ctx, "inventory", params['expression'])
    if ret == None:
        ret = do_query(ctx, kwargs['cache_time'], params['expression'],
                       kwargs['inventory'], kwargs['limit'])
        tmp = {
            'result': ret,
            'expires':
            datetime.utcnow() + timedelta(seconds=kwargs['cache_time'])
        }
        set_cache(ctx, "inventory", params['expression'], tmp)
    else:
        ret = ret['result']

    if kwargs['pager']:
        click.echo_via_pager(
            highlight(
                json.dumps(ret, indent=4), JsonLexer(), TerminalFormatter()))

    return ret
