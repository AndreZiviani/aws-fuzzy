#!/usr/bin/env python3
import sys
import os
from os.path import expanduser
from pathlib import Path
from pprint import pformat
import click
from pygments import highlight
from pygments.lexers import PythonLexer
from pygments.formatters import TerminalFormatter

CONTEXT_SETTINGS = dict(
    help_option_names=['-h', '--help'], auto_envvar_prefix="AWSFUZZY")

VERSION = open(
    os.path.join(os.path.dirname(os.path.abspath(__file__)),
                 'VERSION')).read().strip()


class Environment():
    def __init__(self):
        self.verbose = False

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
        except ImportError as e:
            print(e)
            return None
        return mod.cli


@click.command(cls=ComplexCLI, context_settings=CONTEXT_SETTINGS)
@click.option("-v", "--verbose", is_flag=True, help="Enables verbose mode.")
@click.option(
    "--cache-dir",
    default=expanduser("~") + "/.aws-fuzzy",
    show_default=True,
    help="Cache directory.")
@pass_environment
@click.version_option(version=VERSION)
def cli(ctx, verbose, cache_dir):
    ctx.verbose = verbose
    ctx.cache_dir = cache_dir
    Path(cache_dir).mkdir(parents=True, exist_ok=True)


if __name__ == '__main__':
    cli()
