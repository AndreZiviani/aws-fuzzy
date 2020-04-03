from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params
from .common import cache_params
from .common import query_params
import click


@click.group("inventory")
@click.pass_context
def cli(ctx, **kwargs):
    """Get all resources from AWS service"""


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def acm(ctx, **kwargs):
    """AWS Certificate Manager (ACM) resources"""
    kwargs['service'] = "AWS::ACM::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def apigw(ctx, **kwargs):
    """API Gateway (APIGW) resources"""
    kwargs['service'] = "AWS::ApiGateway::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def apigwv2(ctx, **kwargs):
    """API Gateway V2 (APIGW V2) resources"""
    kwargs['service'] = "AWS::ApiGatewayV2::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def asg(ctx, **kwargs):
    """Auto Scaling Groups (ASG) resources"""
    kwargs['service'] = "AWS::AutoScaling::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def cloudformation(ctx, **kwargs):
    """CloudFormation resources"""
    kwargs['service'] = "AWS::CloudFormation::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def cf(ctx, **kwargs):
    """CloudFront (CF) resources"""
    kwargs['service'] = "AWS::CloudFront::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def ct(ctx, **kwargs):
    """CloudTrail (CT) resources"""
    kwargs['service'] = "AWS::CloudTrail::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def cw(ctx, **kwargs):
    """CloudWatch (CW) resources"""
    kwargs['service'] = "AWS::CloudWatch::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def codebuild(ctx, **kwargs):
    """CodeBuild resources"""
    kwargs['service'] = "AWS::CodeBuild::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def codepipeline(ctx, **kwargs):
    """CodePipeline resources"""
    kwargs['service'] = "AWS::CodePipeline::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def config(ctx, **kwargs):
    """Config resources"""
    kwargs['service'] = "AWS::Config::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def dynamodb(ctx, **kwargs):
    """DynamoDB resources"""
    kwargs['service'] = "AWS::DynamoDB::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
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
        kwargs['service'] = f"AWS::EC2::{kwargs['type']}%"
    else:
        kwargs['service'] = "AWS::EC2::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def eb(ctx, **kwargs):
    """ElasticBeanstalk (EB) resources"""
    kwargs['service'] = "AWS::ElasticBeanstalk::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def elb(ctx, **kwargs):
    """ElasticLoadBalancing (ELB) resources"""
    kwargs['service'] = "AWS::ElasticLoadBalancing::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def elbv2(ctx, **kwargs):
    """ElasticLoadBalancing V2 (ELB) resources"""
    kwargs['service'] = "AWS::ElasticLoadBalancingV2::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def iam(ctx, **kwargs):
    """Identity and Access Management (IAM) resources"""
    kwargs['service'] = "AWS::IAM::%"
    query(ctx, kwargs)


@cli.command(name="lambda")  # lambda is a reserved name in python
@common_params()
@cache_params()
@query_params()
@pass_environment
def awslambda(ctx, **kwargs):
    """Lambda resources"""
    kwargs['service'] = "AWS::Lambda::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def rds(ctx, **kwargs):
    """Relational Database Service (RDS) resources"""
    kwargs['service'] = "AWS::RDS::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def redshift(ctx, **kwargs):
    """Redshift resources"""
    kwargs['service'] = "AWS::Redshift::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def s3(ctx, **kwargs):
    """Simple Storage Service (S3) resources"""
    kwargs['service'] = "AWS::S3::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def ssm(ctx, **kwargs):
    """Systems Manager (SSM) resources"""
    kwargs['service'] = "AWS::SSM::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def servicecatalog(ctx, **kwargs):
    """Service Catalog resources"""
    kwargs['service'] = "AWS::ServiceCatalog::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def shield(ctx, **kwargs):
    """Shield resources"""
    kwargs['service'] = "AWS::Shield::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def shieldr(ctx, **kwargs):
    """Shield Regional resources"""
    kwargs['service'] = "AWS::ShieldRegional::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def waf(ctx, **kwargs):
    """WAF resources"""
    kwargs['service'] = "AWS::WAF::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def wafr(ctx, **kwargs):
    """WAF Regional resources"""
    kwargs['service'] = "AWS::WAFRegional::%"
    query(ctx, kwargs)


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def xray(ctx, **kwargs):
    """XRay resources"""
    kwargs['service'] = "AWS::XRay::%"
    query(ctx, kwargs)
