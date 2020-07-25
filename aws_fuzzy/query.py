import os
import json
import sys

from datetime import timedelta
from datetime import datetime
from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.formatters import TerminalFormatter

import click
import boto3
from aws_fuzzy import common


class Query(common.Cache):
    def __init__(self,
                 ctx,
                 Service=None,
                 Select=None,
                 Filter=None,
                 Limit=None,
                 Aggregator=None,
                 Account=None,
                 Region=None,
                 Pager=True,
                 Cache_time=3600,
                 Profile=None):
        super().__init__(ctx, "inventory", Cache_time)
        self.pager = Pager
        self.region = Region
        self.set_account(Account)

        self.service = Service
        self.limit = Limit

        ctx.vlog(f'Using profile: {Profile}')
        sso = common.SSO(ctx, True, Profile, Cache_time)
        self.client = boto3.Session(
            aws_access_key_id=sso.access_key,
            aws_secret_access_key=sso.secret_key,
            aws_session_token=sso.session_token,
            profile_name=Profile).client('config')

        self.filter = f"resourceType like '{self.service}'"

        if Select:
            self.select = Select
        else:
            self.select = "resourceId, accountId, awsRegion, configuration, tags"

        if Filter:
            self.filter += f" AND {Filter}"
        else:
            if self.profile != 'all':
                self.filter += f" AND accountId like '{self.account_id}'"
            if self.region != 'all':
                self.filter += f" AND awsRegion like '{self.region}'"

        self.expression = f"SELECT {self.select} WHERE {self.filter}"

        ctx.vlog(self.expression)
        if Aggregator is None:
            ret = self.get_cache(os.getenv('AWS_PROFILE', "unknown"))
            if ret is not None:
                self.aggregator = ret['inventory']
            else:
                aggs = self.client.describe_configuration_aggregators()
                try:
                    self.aggregator = aggs['ConfigurationAggregators'][0][
                        'ConfigurationAggregatorName']
                except IndexError:
                    raise Exception(
                        "Could not find any Configuration Aggregator")

                expires = datetime.utcnow() + timedelta(
                    seconds=self.cache_time)
                self.set_cache(
                    os.getenv('AWS_PROFILE', "unknown"), {
                        'inventory': self.aggregator,
                        'expires': expires
                    })

        # TODO:
        # - Add account id to key when caching, or else we dont know which cache to return
        c = self.get_cache(self.expression)
        if c is None:
            self.valid = False
            self.cached = None
        else:
            self.valid = True
            self.cached = c['result']

    def print(self, Pager=None):
        if Pager is None:
            click.echo(
                highlight(
                    json.dumps(self.cached, indent=4), JsonLexer(),
                    TerminalFormatter()))
        if Pager and self.valid:
            click.echo_via_pager(
                highlight(
                    json.dumps(self.cached, indent=4), JsonLexer(),
                    TerminalFormatter()))

    def do_query(self,
                 Expression=None,
                 Aggregator=None,
                 Limit=None,
                 NextToken=None):
        if Expression is None:
            Expression = self.expression
        if Aggregator is None:
            Aggregator = self.aggregator
        if Limit is None:
            Limit = self.limit

        if NextToken:
            return self.client.select_aggregate_resource_config(
                Expression=Expression,
                ConfigurationAggregatorName=Aggregator,
                Limit=Limit,
                NextToken=NextToken)
        return self.client.select_aggregate_resource_config(
            Expression=Expression,
            ConfigurationAggregatorName=Aggregator,
            Limit=Limit)

    def query(self):

        if self.limit <= 100 and self.limit > 0:
            o = self.do_query()
            t = len(o['Results'])
        else:  # Iterate through pages until Limit is reached or end of results
            if self.limit > 0:
                it = self.limit / 100
                mod = self.limit % 100
            else:
                it = sys.maxsize
                mod = 0

            o = self.do_query(Limit=100)

            tmp = o
            i = 1
            r = len(o['Results'])
            t = r
            self.ctx.vlog(f'Got {r} results')
            while 'NextToken' in tmp and i < it:
                tmp = self.do_query(Limit=100, NextToken=tmp['NextToken'])

                o['Results'].extend(tmp['Results'])
                i += 1
                r = len(tmp['Results'])
                t += r
                self.ctx.vlog(f'Got {r} results')

            if mod > 0:
                tmp = self.do_query(Limit=mod, NextToken=tmp['NextToken'])

                o['Results'].extend(tmp['Results'])
                r = len(tmp['Results'])
                t += r
                self.ctx.vlog(f'Got {r} results')

        self.ctx.vlog(f'Got a total of {t} results')

        j = [json.loads(r) for r in o['Results']]

        self.valid = True
        self.cached = j

        expires = datetime.utcnow() + timedelta(seconds=self.cache_time)
        self.set_cache(self.expression, {'result': j, 'expires': expires})
        return j
