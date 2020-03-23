#!/usr/bin/env python3
import boto3
import click
import json
import subprocess
import sys
import functools
from pygments import highlight
from pygments.lexers import JsonLexer
from pygments.formatters import TerminalFormatter

CONTEXT_SETTINGS = dict(help_option_names=['-h', '--help'])

log_level = "normal"


@click.group(context_settings=CONTEXT_SETTINGS)
@click.version_option(version='1.0.0')
def main():
    pass


def debug_log(msg):
    if log_level == "debug":
        print(msg)


def common_params(func):
    @click.option(
        '-d/-nd', '--debug/--no-debug', default=False, help='Enable debug')
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
    @functools.wraps(func)
    def wrapper(*args, **kwargs):
        return func(*args, **kwargs)

    return wrapper


def query(**kwargs):
    global log_level
    if kwargs['debug']:
        log_level = 'debug'

    if 'select' not in kwargs:
        kwargs['select'] = "resourceId, accountId, configuration, tags"

    if len(kwargs['filter']) > 0:
        kwargs[
            'filter'] = f"resourceType like '{kwargs['service']}' AND {kwargs['filter']}"
    else:
        kwargs['filter'] = f"resourceType like '{kwargs['service']}'"

    kwargs[
        'expression'] = f"SELECT {kwargs['select']} WHERE {kwargs['filter']}"

    debug_log(kwargs)

    c = boto3.client('config')

    o = c.select_aggregate_resource_config(
        Expression=kwargs['expression'],
        ConfigurationAggregatorName='linx-digital-inventory-assets',
        Limit=100)

    tmp = o
    while 'NextToken' in tmp:
        tmp = c.select_aggregate_resource_config(
            Expression=kwargs['expression'],
            ConfigurationAggregatorName='linx-digital-inventory-assets',
            Limit=100,
            NextToken=tmp['NextToken'])

        o['Results'].extend(tmp['Results'])

    j = [json.loads(r) for r in o['Results']]

    pager = subprocess.Popen(['less', '-R', '-X', '-K'],
                             stdin=subprocess.PIPE,
                             stdout=sys.stdout)
    pager.stdin.write(
        highlight(json.dumps(j, indent=4), JsonLexer(),
                  TerminalFormatter()).encode())
    pager.stdin.close()
    pager.wait()


@main.command()
@common_params
def acm(**kwargs):
    """ Get all AWS Certificate Manager (ACM) resources """
    kwargs['service'] = "AWS::ACM::Certificate"
    query(**kwargs)


@main.command()
@common_params
def apigw(**kwargs):
    """ Get all API Gateway (APIGW) resources """
    kwargs['service'] = "AWS::ApiGateway::%"
    query(**kwargs)


@main.command()
@common_params
def apigwv2(**kwargs):
    """ Get all API Gateway V2 (APIGW V2) resources """
    kwargs['service'] = "AWS::ApiGatewayV2::%"
    query(**kwargs)


@main.command()
@common_params
def asg(**kwargs):
    """ Get all Auto Scaling Groups (ASG) resources """
    kwargs['service'] = "AWS::AutoScaling::%"
    query(**kwargs)


@main.command()
@common_params
def cf(**kwargs):
    """ Get all CloudFront (CF) resources """
    kwargs['service'] = "AWS::CloudFront::%"
    query(**kwargs)


@main.command()
@common_params
def dynamodb(**kwargs):
    """ Get all DynamoDB resources """
    kwargs['service'] = "AWS::DynamoDB::%"
    query(**kwargs)


@main.command()
@common_params
def ec2(**kwargs):
    """ Get all Elastic Compute Cloud (EC2) resources """
    kwargs['service'] = "AWS::EC2::%"
    query(**kwargs)


@main.command()
@common_params
def iam(**kwargs):
    """ Get all Identity and Access Management (IAM) resources """
    kwargs['service'] = "AWS::IAM::%"
    query(**kwargs)


@main.command(name="lambda")  # lambda is a reserved name in python
@common_params
def awslambda(**kwargs):
    """ Get all Lambda resources """
    kwargs['service'] = "AWS::Lambda::%"
    query(**kwargs)


@main.command()
@common_params
def rds(**kwargs):
    """ Get all Relational Database Service (RDS) resources """
    kwargs['service'] = "AWS::RDS::%"
    query(**kwargs)


@main.command()
@common_params
def s3(**kwargs):
    """ Get all Simple Storage Service (S3) resources """
    kwargs['service'] = "AWS::S3::%"
    query(**kwargs)


if __name__ == '__main__':
    main()
