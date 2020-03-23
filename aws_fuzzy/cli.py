#!/usr/bin/env python3
import boto3
import click
import json
import subprocess
import sys
import functools
import os
from pprint import pformat
from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.lexers import PythonLexer
from pygments.formatters import TerminalFormatter

CONTEXT_SETTINGS = dict(help_option_names=['-h', '--help'])


class Environment(object):
    def __init__(self):
        self.verbose = False
        self.home = os.getcwd()

    def log(self, msg, *args):
        """Logs a message to stderr."""
        if args:
            msg = msg.format(args)
        click.echo(
            highlight(pformat(msg), PythonLexer(), TerminalFormatter()),
            file=sys.stderr)

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
@click.version_option(version='0.0.1')
def cli(ctx, verbose):
    ctx.verbose = verbose


if __name__ == '__main__':
    cli()
