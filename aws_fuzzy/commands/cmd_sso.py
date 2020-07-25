import os

from os.path import expanduser

import click

from aws_fuzzy.cli import pass_environment
from aws_fuzzy import common


@click.group("sso")
@click.pass_context
def cli(ctx, **kwargs):
    """AWS SSO service"""


@cli.command()
@common.p_profile()
@common.p_cache()
@common.p_cache_time()
@pass_environment
def login(ctx, **kwargs):
    """Login to AWS SSO"""

    sso = common.SSO(
        ctx,
        Cache=kwargs['cache'],
        Account=kwargs['profile'],
        Cache_time=kwargs['cache_time'])

    ctx.vlog(kwargs)
    if sso.valid:
        # We got valid cached credentials
        sso.print_credentials()
        return

    # Missing or expired cached credentials, requesting new one
    ctx.vlog("Could not find cached credentials or they are expired")
    sso.get_new_credentials()
    sso.print_credentials()


@cli.command()
@common.p_region(region='us-east-1')
@common.p_cache()
@common.p_cache_time()
@pass_environment
def configure(ctx, **kwargs):
    """Configure AWS SSO"""

    sso_url = None
    sso_profiles = expanduser('~') + '/.aws/config'
    sso_region = kwargs['region']

    if os.path.isfile(sso_profiles):
        ctx.log(
            "AWS config file already exists, making a backup before updating")
        p = common.Common(ctx)

        if 'default' in p.profiles:
            sso_url = p.profiles['default']['sso_start_url']
            sso_region = p.profiles['default']['sso_region']

        os.rename(sso_profiles, f"{sso_profiles}.bkp")

    start = click.prompt("Enter SSO start url", default=sso_url)
    region = click.prompt("Default region", default=sso_region)

    f = open(sso_profiles, 'w')
    f.write(f"""[default]
sso_start_url = {start}
sso_region = {region}
sso_role_name = dummy
sso_account_id = 00000000""")
    f.close()

    sso = common.SSO(
        ctx,
        Cache=kwargs['cache'],
        Account='default',
        Cache_time=kwargs['cache_time'])

    accounts = sso.list_accounts(region=sso_region, profile='default')

    f = open(sso_profiles, 'a')
    for account in accounts:
        f.write(f"""
[profile {account['name']}]
sso_start_url = {start}
sso_region = {region}
sso_account_id = {account['id']}
sso_role_name = {account['role']}""")

    f.close()

    ctx.log("Done! Please check ~/.aws/config for your profiles")
