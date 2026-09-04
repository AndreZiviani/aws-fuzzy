# aws-fuzzy
[![Coverage Status](https://coveralls.io/repos/github/AndreZiviani/aws-fuzzy/badge.svg?branch=master)](https://coveralls.io/github/AndreZiviani/aws-fuzzy?branch=master)
[![Actions](https://github.com/AndreZiviani/aws-fuzzy/actions/workflows/release.yml/badge.svg)](https://github.com/AndreZiviani/aws-fuzzy/actions)

**aws-fuzzy was previously developed in Python, old code is available [here](https://github.com/AndreZiviani/aws-fuzzy/tree/legacy-python)**

aws-fuzzy is a tool to retrieve information from multiple AWS services.

- **AWS Config**: Retrieve inventory information from AWS Config
- **Chart**: Plot the connection between different AWS components (e.g. VPC Peering)
- **SSH**: Search EC2 instances using [fuzzy finder](https://github.com/junegunn/fzf)
- **SSM**: Open a shell, forward a port or run a command on an EC2 instance via SSM Session Manager
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
  ssm     Interact with EC2 instances via SSM
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
  nm        Chart NetworkManager topology
  peering   Chart peering relationship
  tgroutes  Chart TransitGateway route tables
```

A graph is generated in HTML format, with account name based on profiles defined AWS config file.
The HTML file containing the graph is saved in the current directory.

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

## SSM

Interact with EC2 instances through SSM Session Manager, without needing SSH
access or a bastion host. Only instances that are running, registered with SSM
and reachable are listed.

```sh
Usage:
  aws-fuzzy [OPTIONS] ssm <command>

Interact with EC2 instances via SSM

Available commands:
  session      Start a session on a EC2 instance
  portforward  Start a portforwarding session on a EC2 instance
  run          Run a command on a EC2 instance and print its output
```

Every subcommand accepts the same target options:

```sh
      -p, --profile=          What profile to use (default: $AWS_PROFILE) [$AWSFUZZY_PROFILE, $AWS_PROFILE]
      -r, --region=           What AWS region to use (default: us-east-1) [$AWS_REGION, $AWS_DEFAULT_REGION]
      -i, --instance=         Target instance, by instance id, Name tag or private IP
          --non-interactive   Never open the instance picker, fail instead if the target is
                              missing or ambiguous
```

Omit `-i` and you get the fuzzy picker, as before. Pass it and the instance is
resolved directly:

```sh
# by instance id, Name tag or private IP
aws-fuzzy ssm session -i i-0123456789abcdef0
aws-fuzzy ssm session -i web-prod-1
aws-fuzzy ssm session -i 10.0.1.5
```

A query is matched against the instance id, then the private IP, then the
`Name` tag exactly, then the `Name` tag as a substring — the first of those
that hits anything wins. If a query matches several instances the picker opens
containing just those candidates; with `--non-interactive` it fails instead and
lists their instance ids.

### run

`ssm run` sends a single shell command to one instance, waits for it, prints
its output and exits with the remote exit code, so it composes with other
commands in a script:

```sh
$ aws-fuzzy ssm run -i web-prod-1 -c 'systemctl is-active nginx'
active
$ echo $?
0

$ aws-fuzzy ssm run -i web-prod-1 -c 'exit 3'; echo $?
3
```

```sh
      -c, --command=  Command to run on the remote instance (required)
          --timeout=  How long to wait for the command to finish, in seconds (default: 60)
```

A command that exceeds `--timeout` exits `124`, matching `timeout(1)`. SSM
truncates command output at 24000 bytes; when that happens `run` prints what it
received and warns on stderr.

## SSO

Configure and login to AWS SSO and export session credentials.
This feature uses the awesome [Granted](https://github.com/common-fate/granted) under the hood.

```sh
Usage:
  aws-fuzzy [OPTIONS] sso <command>

Utilities developed to ease operation and configuration of AWS SSO.
This is mostly imported from common-fate/granted so some log messages may display 'granted' as the application name

Help Options:
  -h, --help      Show this help message

Available commands:
  browser    Configure default browser
  configure  Configure AWS SSO
  console    Open AWS Console
  login      Login to AWS
```
