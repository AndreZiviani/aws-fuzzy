import click
from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import Query
from aws_fuzzy import common


@click.group("inventory")
@click.pass_context
def cli(ctx, **kwargs):
    """Get all resources from AWS Config service"""


def do_query(ctx, kwargs):
    query = Query(
        ctx,
        Service=kwargs['service'],
        Select=kwargs['select'],
        Filter=kwargs['filter'],
        Limit=kwargs['limit'],
        Account=kwargs['account'],
        Region=kwargs['region'],
        Pager=kwargs['pager'],
        Cache_time=kwargs['cache_time'],
        Profile=kwargs['profile'])

    if query.valid:
        query.print()
    else:
        query.query()
        query.print()


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def acm(ctx, **kwargs):
    """AWS Certificate Manager (ACM) resources"""
    kwargs['service'] = "AWS::ACM::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def apigw(ctx, **kwargs):
    """API Gateway (APIGW) resources"""
    kwargs['service'] = "AWS::ApiGateway::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def apigwv2(ctx, **kwargs):
    """API Gateway V2 (APIGW V2) resources"""
    kwargs['service'] = "AWS::ApiGatewayV2::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def asg(ctx, **kwargs):
    """Auto Scaling Groups (ASG) resources"""
    kwargs['service'] = "AWS::AutoScaling::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def cloudformation(ctx, **kwargs):
    """CloudFormation resources"""
    kwargs['service'] = "AWS::CloudFormation::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def cf(ctx, **kwargs):
    """CloudFront (CF) resources"""
    kwargs['service'] = "AWS::CloudFront::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def ct(ctx, **kwargs):
    """CloudTrail (CT) resources"""
    kwargs['service'] = "AWS::CloudTrail::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def cw(ctx, **kwargs):
    """CloudWatch (CW) resources"""
    kwargs['service'] = "AWS::CloudWatch::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def codebuild(ctx, **kwargs):
    """CodeBuild resources"""
    kwargs['service'] = "AWS::CodeBuild::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def codepipeline(ctx, **kwargs):
    """CodePipeline resources"""
    kwargs['service'] = "AWS::CodePipeline::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def config(ctx, **kwargs):
    """Config resources"""
    kwargs['service'] = "AWS::Config::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def dynamodb(ctx, **kwargs):
    """DynamoDB resources"""
    kwargs['service'] = "AWS::DynamoDB::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@click.option(
    '-t',
    '--type',
    default='all',
    show_default='all',
    show_envvar=True,
    help='Filter by EC2 resource (case sensitive): ['
    'CustomerGateway, '
    'EIP, '
    'Host, '
    'Instance, '
    'InternetGateway, '
    'NetworkAcl, '
    'NetworkInterface, '
    'RegisteredHAInstance, '
    'RouteTable, '
    'SecurityGroup, '
    'Subnet, '
    'VPC, '
    'VPCEndpoint, '
    'VPCEndpointService, '
    'VPCPeeringConnection, '
    'VPNConnection, '
    'VPNGateway, '
    'Volume'
    ']')
@pass_environment
def ec2(ctx, **kwargs):
    """Elastic Compute Cloud (EC2) resources"""
    if kwargs['type'] != 'all':
        kwargs['service'] = f"AWS::EC2::{kwargs['type']}"
    else:
        kwargs['service'] = "AWS::EC2::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def eb(ctx, **kwargs):
    """ElasticBeanstalk (EB) resources"""
    kwargs['service'] = "AWS::ElasticBeanstalk::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def elb(ctx, **kwargs):
    """ElasticLoadBalancing (ELB) resources"""
    kwargs['service'] = "AWS::ElasticLoadBalancing::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def elbv2(ctx, **kwargs):
    """ElasticLoadBalancing V2 (ELB) resources"""
    kwargs['service'] = "AWS::ElasticLoadBalancingV2::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def iam(ctx, **kwargs):
    """Identity and Access Management (IAM) resources"""
    kwargs['service'] = "AWS::IAM::%"
    do_query(ctx, kwargs)


@cli.command(name="lambda")  # lambda is a reserved name in python
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def awslambda(ctx, **kwargs):
    """Lambda resources"""
    kwargs['service'] = "AWS::Lambda::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def rds(ctx, **kwargs):
    """Relational Database Service (RDS) resources"""
    kwargs['service'] = "AWS::RDS::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def redshift(ctx, **kwargs):
    """Redshift resources"""
    kwargs['service'] = "AWS::Redshift::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def s3(ctx, **kwargs):
    """Simple Storage Service (S3) resources"""
    kwargs['service'] = "AWS::S3::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def ssm(ctx, **kwargs):
    """Systems Manager (SSM) resources"""
    kwargs['service'] = "AWS::SSM::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def servicecatalog(ctx, **kwargs):
    """Service Catalog resources"""
    kwargs['service'] = "AWS::ServiceCatalog::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def shield(ctx, **kwargs):
    """Shield resources"""
    kwargs['service'] = "AWS::Shield::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def shieldr(ctx, **kwargs):
    """Shield Regional resources"""
    kwargs['service'] = "AWS::ShieldRegional::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def waf(ctx, **kwargs):
    """WAF resources"""
    kwargs['service'] = "AWS::WAF::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def wafr(ctx, **kwargs):
    """WAF Regional resources"""
    kwargs['service'] = "AWS::WAFRegional::%"
    do_query(ctx, kwargs)


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def xray(ctx, **kwargs):
    """XRay resources"""
    kwargs['service'] = "AWS::XRay::%"
    do_query(ctx, kwargs)
