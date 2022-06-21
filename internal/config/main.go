package config

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
)

type Ec2Config struct {
	Type string `short:"t" long:"type" default:"Instance" description:"Filter by EC2 resource (case sensitive):\n CustomerGateway, EgressOnlyInternetGateway, EIP, FlowLog, Host, Instance, InternetGateway, NatGateway, NetworkAcl, NetworkInterface, RegisteredHAInstance, RouteTable, SecurityGroup, Subnet, Volume, VPCEndpoint, VPCEndpointService, VPCPeeringConnection, VPC, VPNConnection, VPNGateway"`
	Config
}
type Config struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Pager   bool   `long:"pager" description:"Pipe output to less"`
	Account string `short:"a" long:"account" description:"Filter Config resources to this account"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" description:"What region to use, if not specified defaults to $AWS_DEFAULT_REGION or us-east-1"`
	Select  string `short:"s" long:"select" default:"resourceId, accountId, awsRegion, configuration, tags" description:"Custom select to filter results"`
	Filter  string `short:"f" long:"filter" description:"Use a custom query to filter results"`
	Limit   int    `short:"l" long:"limit" default:"0" description:"Limit the number of results"`
	Service string
}

var AwsServices = map[string]string{
	"acm":            "ACM",
	"apigw":          "ApiGateway",
	"apigwv2":        "ApiGatewayV2",
	"asg":            "AutoScaling",
	"cfn":            "CloudFormation",
	"cf":             "CloudFront",
	"ct":             "CloudTrail",
	"cw":             "CloudWatch",
	"cb":             "CodeBuild",
	"cp":             "CodePipeline",
	"config":         "Config",
	"dynamo":         "DynamoDB",
	"ec2":            "EC2",
	"eb":             "ElasticBeanstalk",
	"elb":            "ElasticLoadBalancing",
	"elbv2":          "ElasticLoadBalancingV2",
	"es":             "Elasticsearch",
	"iam":            "IAM",
	"kms":            "KMS",
	"lambda":         "Lambda",
	"qldb":           "QLDB",
	"rds":            "RDS",
	"redshift":       "Redshift",
	"s3":             "S3",
	"servicecatalog": "ServiceCatalog",
	"shield":         "Shield",
	"shieldreg":      "ShieldRegional",
	"sns":            "SNS",
	"sqs":            "SQS",
	"ssm":            "SSM",
	"waf":            "WAF",
	"wafreg":         "WAFRegional",
	"wafv2":          "WAFv2",
	"xray":           "XRay",
}

func Init(parser *flags.Parser) {
	cmd, err := parser.AddCommand(
		"config",
		"Interact with AWS Config inventory",
		"Interact with AWS Config inventory",
		&struct{}{})

	if err != nil {
		return
	}

	for k, v := range AwsServices {
		if k == "ec2" {
			config := Ec2Config{}
			config.Service = v
			cmd.AddCommand(
				k,
				fmt.Sprintf("Query %s resources in AWS Config inventory", v),
				fmt.Sprintf("Query %s resources in AWS Config inventory", v),
				&config)
			continue
		}
		config := Config{}
		config.Service = v
		cmd.AddCommand(
			k,
			fmt.Sprintf("Query %s resources in AWS Config inventory", v),
			fmt.Sprintf("Query %s resources in AWS Config inventory", v),
			&config)
	}

}
