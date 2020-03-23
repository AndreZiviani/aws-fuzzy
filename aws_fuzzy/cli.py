#!/usr/bin/env python3
import boto3
import click
import json
import subprocess
import sys
import functools
import os
from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.formatters import TerminalFormatter

CONTEXT_SETTINGS = dict(help_option_names=['-h', '--help'])


class Environment(object):
    def __init__(self):
        self.verbose = False
        self.home = os.getcwd()

    def log(self, msg, *args):
        """Logs a message to stderr."""
        if args:
            msg %= args
        click.echo(msg, file=sys.stderr)

    def vlog(self, msg, *args):
        """Logs a message to stderr only if verbose is enabled."""
        if self.verbose:
            self.log(msg, *args)


pass_environment = click.make_pass_decorator(Environment, ensure=True)
cmd_folder = os.path.abspath(
    os.path.join(os.path.dirname(__file__), "commands"))


class ComplexCLI(click.MultiCommand):
    def list_commands(self, ctx):
        rv = []
        for filename in os.listdir(cmd_folder):
            if filename.endswith(".py") and filename.startswith("cmd_"):
                rv.append(filename[4:-3])
        rv.sort()
        return rv

    def get_command(self, ctx, name):
        try:
            if sys.version_info[0] == 2:
                name = name.encode("ascii", "replace")
            mod = __import__("aws_fuzzy.commands.cmd_{}".format(name), None,
                             None, ["cli"])
        except ImportError:
            print('err')
            return
        return mod.cli


@click.command(cls=ComplexCLI, context_settings=CONTEXT_SETTINGS)
@click.option("-v", "--verbose", is_flag=True, help="Enables verbose mode.")
@pass_environment
@click.version_option(version='1.0.0')
def cli(ctx, verbose):
    ctx.verbose = verbose


def query(ctx, **kwargs):

    if 'select' not in kwargs:
        kwargs['select'] = "resourceId, accountId, configuration, tags"

    if len(kwargs['filter']) > 0:
        kwargs[
            'filter'] = f"resourceType like '{kwargs['service']}' AND {kwargs['filter']}"
    else:
        kwargs['filter'] = f"resourceType like '{kwargs['service']}'"

    kwargs[
        'expression'] = f"SELECT {kwargs['select']} WHERE {kwargs['filter']}"

    ctx.vlog(kwargs)

    c = boto3.client('config')

    if kwargs['limit'] <= 100:
        o = c.select_aggregate_resource_config(
            Expression=kwargs['expression'],
            ConfigurationAggregatorName='linx-digital-inventory-assets',
            Limit=kwargs['limit'])
    else:
        it = kwargs['limit'] / 100
        mod = kwargs['limit'] % 100

        o = c.select_aggregate_resource_config(
            Expression=kwargs['expression'],
            ConfigurationAggregatorName='linx-digital-inventory-assets',
            Limit=100)

        tmp = o
        i = 1
        while 'NextToken' in tmp and i < it:
            tmp = c.select_aggregate_resource_config(
                Expression=kwargs['expression'],
                ConfigurationAggregatorName='linx-digital-inventory-assets',
                Limit=100,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])

        if mod > 0:
            tmp = c.select_aggregate_resource_config(
                Expression=kwargs['expression'],
                ConfigurationAggregatorName='linx-digital-inventory-assets',
                Limit=mod,
                NextToken=tmp['NextToken'])

            o['Results'].extend(tmp['Results'])

    j = [json.loads(r) for r in o['Results']]

    if kwargs['pager']:

        pager = subprocess.Popen(['less', '-R', '-X', '-K'],
                                 stdin=subprocess.PIPE,
                                 stdout=sys.stdout)
        pager.stdin.write(
            highlight(
                json.dumps(j, indent=4), JsonLexer(),
                TerminalFormatter()).encode())
        pager.stdin.close()
        pager.wait()

    return j


if __name__ == '__main__':
    cli()
