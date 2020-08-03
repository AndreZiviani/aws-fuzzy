# aws-fuzzy
[![PyPI version](https://badge.fury.io/py/aws-fuzzy.svg)](https://badge.fury.io/py/aws-fuzzy)
[![PyPI - Downloads](https://img.shields.io/pypi/dm/aws-fuzzy)](https://pypi.org/project/aws-fuzzy/)
[![Coverage Status](https://coveralls.io/repos/github/AndreZiviani/aws-fuzzy/badge.svg?branch=master)](https://coveralls.io/github/AndreZiviani/aws-fuzzy?branch=master)
[![CircleCI](https://circleci.com/gh/AndreZiviani/aws-fuzzy/tree/master.svg?style=shield)](https://circleci.com/gh/AndreZiviani/aws-fuzzy/tree/master)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3879/badge)](https://bestpractices.coreinfrastructure.org/en/projects/3879)

aws-fuzzy is a tool to retrieve information from multiple AWS services.

- **AWS Config**: Retrieve inventory information from AWS Config
- **Plot**: Plot the connection between different AWS components (e.g. VPC Peering)
- **SSH**: Search EC2 instances using [fuzzy finder](https://github.com/junegunn/fzf)
- **SSO**: Login and export AWS credentials as environment variables
- **Cache**: all results can be optionally cached to improve performance


# Install

```sh
$ pip install aws-fuzzy
```


# Usage

aws-fuzzy will query AWS services using the profile specified from environment variable `AWS_PROFILE` or from a runtime parameter `-p`, the credentials are retrieved from environment variables or from `~/.aws/credentials`.
Each option can be specified via environment variables.

```sh
Usage: aws-fuzzy [OPTIONS] COMMAND [ARGS]...

Options:
  -v, --verbose     Enables verbose mode.
  --cache-dir TEXT  Cache directory.  [default: /home/user/.aws-fuzzy]
  --version         Show the version and exit.
  -h, --help        Show this message and exit.

Commands:
  inventory  Get all resources from AWS service
  plot       Plot resources from AWS
  ssh        SSH to EC2 instance
  sso        AWS SSO service
```

## Invetory

The `inventory` module queries AWS Config service (with an optional `aggregator` parameter).

```sh
Usage: aws-fuzzy inventory [OPTIONS] COMMAND [ARGS]...

  Get all resources from AWS service

Options:
  -h, --help  Show this message and exit.

Commands:
  acm             AWS Certificate Manager (ACM) resources
  apigw           API Gateway (APIGW) resources
  apigwv2         API Gateway V2 (APIGW V2) resources
  asg             Auto Scaling Groups (ASG) resources
  cf              CloudFront (CF) resources
  cloudformation  CloudFormation resources
  codebuild       CodeBuild resources
  codepipeline    CodePipeline resources
  config          Config resources
  ct              CloudTrail (CT) resources
  cw              CloudWatch (CW) resources
  dynamodb        DynamoDB resources
  eb              ElasticBeanstalk (EB) resources
  ec2             Elastic Compute Cloud (EC2) resources
  elb             ElasticLoadBalancing (ELB) resources
  elbv2           ElasticLoadBalancing V2 (ELB) resources
  iam             Identity and Access Management (IAM) resources
  lambda          Lambda resources
  rds             Relational Database Service (RDS) resources
  redshift        Redshift resources
  s3              Simple Storage Service (S3) resources
  servicecatalog  Service Catalog resources
  shield          Shield resources
  shieldr         Shield Regional resources
  ssm             Systems Manager (SSM) resources
  waf             WAF resources
  wafr            WAF Regional resources
  xray            XRay resources
```

Each module can have different options, `ec2` for example have additional options due to its large number of resources

```sh
Usage: aws-fuzzy inventory ec2 [OPTIONS]

  Elastic Compute Cloud (EC2) resources

Options:
  -a, --account TEXT    Filter by accountid  [env var:
                        AWSFUZZY_INVENTORY_EC2_ACCOUNT; default: all]

  -s, --select TEXT     Custom select to filter results  [env var:
                        AWSFUZZY_INVENTORY_EC2_SELECT]

  -r, --region TEXT     Filter by region  [env var:
                        AWSFUZZY_INVENTORY_EC2_REGION; default: all]

  -f, --filter TEXT     Use a custom query to filter results  [env var:
                        AWSFUZZY_INVENTORY_EC2_FILTER; default: '']

  --pager / --no-pager  Send query results to pager  [env var:
                        AWSFUZZY_INVENTORY_EC2_PAGER; default: True]

  -l, --limit INTEGER   Use a custom query to filter results  [env var:
                        AWSFUZZY_INVENTORY_EC2_LIMIT; default: 0]

  --cache / --no-cache  Whether to use cached results  [env var:
                        AWSFUZZY_INVENTORY_EC2_CACHE; default: True]

  --cache-time INTEGER  Cache results TTL in seconds  [env var:
                        AWSFUZZY_INVENTORY_EC2_CACHE_TIME; default: 3600]

  -i, --inventory TEXT  Cache results TTL in seconds  [env var:
                        AWSFUZZY_INVENTORY_EC2_INVENTORY; default: (First one
                        found)]

  -t, --type TEXT       Filter by EC2 resource (case sensitive):
                        [CustomerGateway, EIP, Host, Instance,
                        InternetGateway, NetworkAcl, NetworkInterface,
                        RegisteredHAInstance, RouteTable, SecurityGroup,
                        Subnet, VPC, VPCEndpoint, VPCEndpointService,
                        VPCPeeringConnection, VPNConnection, VPNGateway,
                        Volume]  [env var: AWSFUZZY_INVENTORY_EC2_TYPE;
                        default: (all)]

  -h, --help            Show this message and exit.
```

## Plot

It can also plot a graph of the relationship between resources.

```sh
Usage: aws-fuzzy plot [OPTIONS] COMMAND [ARGS]...

  Plot resources from AWS

Options:
  -h, --help  Show this message and exit.

Commands:
  vpcpeering  Plot VPC Peering connections graph
  securitygroups  Plot which SecurityGroups have a relationship with the...
```

A graph is generated in HTML format, each node is color coded with account name (based on profiles defined AWS config file) containing the resource, each edge represents a connection and have a tooltip showing the tag (in case of `vpcpeering`) or port (`securitygroup` mode).
The HTML file containing the graph is saved in the current directory with the name `mygraph.html`.

## SSH

Use [fuzzy finder](https://github.com/junegunn/fzf) to select and SSH to instances.

```sh
Usage: aws-fuzzy ssh [OPTIONS]

  SSH to EC2 instance

Options:
  -a, --account TEXT    Filter by accountid  [env var: AWSFUZZY_SSH_ACCOUNT;
                        default: all]

  -s, --select TEXT     Custom select to filter results  [env var:
                        AWSFUZZY_SSH_SELECT]

  -r, --region TEXT     Filter by region  [env var: AWSFUZZY_SSH_REGION;
                        default: all]

  -f, --filter TEXT     Use a custom query to filter results  [env var:
                        AWSFUZZY_SSH_FILTER; default: '']

  --pager / --no-pager  Send query results to pager  [env var:
                        AWSFUZZY_SSH_PAGER; default: False]

  -l, --limit INTEGER   Use a custom query to filter results  [env var:
                        AWSFUZZY_SSH_LIMIT; default: 0]

  -u, --user TEXT       Username to use with SSH  [env var: AWSFUZZY_SSH_USER;
                        default: '']

  -k, --key TEXT        SSH key path  [env var: AWSFUZZY_SSH_KEY; default: '']
  --cache / --no-cache  Whether to use cached results  [env var:
                        AWSFUZZY_SSH_CACHE; default: True]

  --cache-time INTEGER  Cache results TTL in seconds  [env var:
                        AWSFUZZY_SSH_CACHE_TIME; default: 3600]

  -h, --help            Show this message and exit.
```

## SSO

Configure and login to AWS SSO and export session credentials.

```sh
Usage: aws-fuzzy sso [OPTIONS] COMMAND [ARGS]...

  AWS SSO service

Options:
  -h, --help  Show this message and exit.

Commands:
  configure  Configure AWS SSO
  login      Login to AWS SSO
```
