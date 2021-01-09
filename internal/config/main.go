package config

import (
	"fmt"
	flags "github.com/jessevdk/go-flags"
)

type Ec2ConfigCommand struct {
	Type string `short:"t" long:"type" default:"Instance" description:"Filter by EC2 resource (case sensitive):\n CustomerGateway, EgressOnlyInternetGateway, EIP, FlowLog, Host, Instance, InternetGateway, NatGateway, NetworkAcl, NetworkInterface, RegisteredHAInstance, RouteTable, SecurityGroup, Subnet, Volume, VPCEndpoint, VPCEndpointService, VPCPeeringConnection, VPC, VPNConnection, VPNGateway"`
	ConfigCommand
}
type ConfigCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Pager   bool   `long:"pager" description:"Pipe output to less"`
	Account string `short:"a" long:"account" description:"Filter Config resources to this account"`
	Service string
}

/*
var (
	configCommand ConfigCommand
)
*/

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
			configCommand := Ec2ConfigCommand{}
			configCommand.Service = v
			cmd.AddCommand(
				k,
				fmt.Sprintf("Query %s resources in AWS Config inventory", v),
				fmt.Sprintf("Query %s resources in AWS Config inventory", v),
				&configCommand)
			continue
		}
		configCommand := ConfigCommand{}
		configCommand.Service = v
		cmd.AddCommand(
			k,
			fmt.Sprintf("Query %s resources in AWS Config inventory", v),
			fmt.Sprintf("Query %s resources in AWS Config inventory", v),
			&configCommand)
	}

}
