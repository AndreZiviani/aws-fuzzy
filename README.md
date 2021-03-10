# aws-fuzzy
[![Coverage Status](https://coveralls.io/repos/github/AndreZiviani/aws-fuzzy/badge.svg?branch=master)](https://coveralls.io/github/AndreZiviani/aws-fuzzy?branch=master)
[![CircleCI](https://circleci.com/gh/AndreZiviani/aws-fuzzy/tree/master.svg?style=shield)](https://circleci.com/gh/AndreZiviani/aws-fuzzy/tree/master)

** aws-fuzzy was previously developed in Python, old code is available [here](https://github.com/AndreZiviani/aws-fuzzy/tree/legacy-python) **

aws-fuzzy is a tool to retrieve information from multiple AWS services.

- **AWS Config**: Retrieve inventory information from AWS Config
- **Chart**: Plot the connection between different AWS components (e.g. VPC Peering)
- **SSH**: Search EC2 instances using [fuzzy finder](https://github.com/junegunn/fzf)
- **SSO**: Login and export AWS credentials as environment variables
- **Cache**: all results can be optionally cached to improve performance


# Install

Check the releases page [here](https://github.com/AndreZiviani/aws-fuzzy/releases)


# Usage

aws-fuzzy will query AWS services using the profile specified from environment variable `AWS_PROFILE` or from a runtime parameter `-p`, the credentials are retrieved from environment variables or from `~/.aws/credentials`.
Each option can be specified via environment variables.

```sh
Usage:
  aws-fuzzy [OPTIONS] <command>

Help Options:
  -h, --help  Show this help message

Available commands:
  chart   Chart
  config  Interact with AWS Config inventory
  ssh     SSH to EC2 instances
  sso     SSO Utilities
```

## Config

The `config` module queries AWS Config service (with an optional `aggregator` parameter).

```sh
Usage:
  aws-config [OPTIONS] config <command>

Interact with AWS Config inventory

Help Options:
  -h, --help      Show this help message

Available commands:
  acm             Query ACM resources in AWS Config inventory
  apigw           Query ApiGateway resources in AWS Config inventory
  apigwv2         Query ApiGatewayV2 resources in AWS Config inventory
  asg             Query AutoScaling resources in AWS Config inventory
  cb              Query CodeBuild resources in AWS Config inventory
  cf              Query CloudFront resources in AWS Config inventory
  cfn             Query CloudFormation resources in AWS Config inventory
  config          Query Config resources in AWS Config inventory
  cp              Query CodePipeline resources in AWS Config inventory
  ct              Query CloudTrail resources in AWS Config inventory
  cw              Query CloudWatch resources in AWS Config inventory
  dynamo          Query DynamoDB resources in AWS Config inventory
  eb              Query ElasticBeanstalk resources in AWS Config inventory
  ec2             Query EC2 resources in AWS Config inventory
  elb             Query ElasticLoadBalancing resources in AWS Config inventory
  elbv2           Query ElasticLoadBalancingV2 resources in AWS Config inventory
  es              Query Elasticsearch resources in AWS Config inventory
  iam             Query IAM resources in AWS Config inventory
  kms             Query KMS resources in AWS Config inventory
  lambda          Query Lambda resources in AWS Config inventory
  qldb            Query QLDB resources in AWS Config inventory
  rds             Query RDS resources in AWS Config inventory
  redshift        Query Redshift resources in AWS Config inventory
  s3              Query S3 resources in AWS Config inventory
  servicecatalog  Query ServiceCatalog resources in AWS Config inventory
  shield          Query Shield resources in AWS Config inventory
  shieldreg       Query ShieldRegional resources in AWS Config inventory
  sns             Query SNS resources in AWS Config inventory
  sqs             Query SQS resources in AWS Config inventory
  ssm             Query SSM resources in AWS Config inventory
  waf             Query WAF resources in AWS Config inventory
  wafreg          Query WAFRegional resources in AWS Config inventory
  wafv2           Query WAFv2 resources in AWS Config inventory
  xray            Query XRay resources in AWS Config inventory
```

Each module can have different options, `ec2` for example have additional options due to its large number of resources

```sh
Usage:
  aws-fuzzy [OPTIONS] config ec2 [ec2-OPTIONS]

Query EC2 resources in AWS Config inventory

Help Options:
  -h, --help         Show this help message

[ec2 command options]
      -t, --type=    Filter by EC2 resource (case sensitive):
                     CustomerGateway, EgressOnlyInternetGateway, EIP, FlowLog, Host, Instance, InternetGateway, NatGateway, NetworkAcl,
                     NetworkInterface, RegisteredHAInstance, RouteTable, SecurityGroup, Subnet, Volume, VPCEndpoint, VPCEndpointService,
                     VPCPeeringConnection, VPC, VPNConnection, VPNGateway (default: Instance)
      -p, --profile= What profile to use (default: default) [$AWS_PROFILE]
          --pager    Pipe output to less
      -a, --account= Filter Config resources to this account
      -s, --select=  Custom select to filter results (default: resourceId, accountId, awsRegion, configuration, tags)
      -f, --filter=  Use a custom query to filter results
      -l, --limit=   Limit the number of results (default: 0)
```

## Chart

It can also plot a graph of the relationship between resources.

```sh
Usage:
  main [OPTIONS] chart peering

Chart relationship between resources

Help Options:
  -h, --help      Show this help message

Available commands:
  peering  Chart peering relationship
```

A graph is generated in HTML format, each node is color coded with account name (based on profiles defined AWS config file) containing the resource, each edge represents a connection and have a tooltip showing the tag (in case of `vpcpeering`) or port (`securitygroup` mode).
The HTML file containing the graph is saved in the current directory with the name `graph.html`.

## SSH

Use [fuzzy finder](https://github.com/junegunn/fzf) to select and SSH to instances.

```sh
Usage:
  aws-fuzzy [OPTIONS] ssh [ssh-OPTIONS]

SSH to EC2 instances

Help Options:
  -h, --help         Show this help message

[ssh command options]
      -p, --profile= What profile to use (default: default) [$AWS_PROFILE]
      -u, --user=    Username to use with SSH (default: $USER) [$AWSFUZZY_SSH_USER]
      -k, --key=     Key to use with SSH (default: ~/.ssh/id_rsa) [$AWSFUZZY_SSH_KEY]
```

## SSO

Configure and login to AWS SSO and export session credentials.

```sh
Usage:
  aws-fuzzy [OPTIONS] sso <configure | login>

Utilities developed to ease operation and configuration of AWS SSO

Help Options:
  -h, --help      Show this help message

Available commands:
  configure  Configure AWS SSO
  login      Login to AWS SSO
```
