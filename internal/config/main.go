package config

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

type Config struct {
	Profile string
	Pager   bool
	Account string
	Region  string
	Select  string
	Filter  string
	Limit   int
	Type    string
	Service string
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func Command() *cli.Command {
	subcommands := make([]*cli.Command, 0)
	for k, v := range AwsServices {
		subcommands = append(subcommands, &cli.Command{
			Name:  k,
			Usage: fmt.Sprintf("Query %v service", v.Name),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
				&cli.BoolFlag{Name: "pager", Usage: "Pipe output to less", Value: false},
				&cli.StringFlag{Name: "account", Aliases: []string{"a"}, Usage: "Filter Config resources to this account"},
				&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
				&cli.StringFlag{Name: "select", Aliases: []string{"s"}, Usage: "Custom select to filter results", Value: "resourceId, accountId, awsRegion, configuration, tags"},
				&cli.StringFlag{Name: "filter", Aliases: []string{"f"}, Usage: "Complete custom query"},
				&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Usage: "Limit the number of results", Value: 0},
				&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Value: "%", Usage: fmt.Sprintf("Filter results to only one of the following types: %v", v.Types)},
			},
			Action: func(c *cli.Context) error {
				serviceType := c.String("type")
				if serviceType != "%" {
					if service, ok := AwsServices[c.Command.Name]; ok {
						if !contains(service.Types, serviceType) {
							return fmt.Errorf("Could not find type '%s' for service '%s'", serviceType, service.Name)
						}
					}
				}
				config := New(c.String("profile"),
					c.String("account"),
					c.String("region"),
					c.String("select"),
					c.String("filter"),
					AwsServices[c.Command.Name].Name, //service
					serviceType,
					c.Bool("pager"),
					c.Int("limit"),
				)

				return config.Execute(c.Context)
			},
		})
	}
	command := cli.Command{
		Name:        "config",
		Usage:       "Interact with AWS Config inventory",
		Subcommands: subcommands,
	}

	return &command
}

type AwsService struct {
	Name  string
	Types []string
}

var AwsServices = map[string]AwsService{
	"acm":          AwsService{Name: "ACM", Types: []string{"Certificate"}},
	"mq":           AwsService{Name: "AmazonMQ", Types: []string{"Broker"}},
	"apigw":        AwsService{Name: "ApiGateway", Types: []string{"RestApi", "Stage"}},
	"apigwv2":      AwsService{Name: "ApiGatewayV2", Types: []string{"Api", "Stage"}},
	"appconfig":    AwsService{Name: "AppConfig", Types: []string{"Application", "ConfigurationProfile", "Environment"}},
	"asg":          AwsService{Name: "AutoScaling", Types: []string{"AutoScalingGroup", "LaunchConfiguration", "ScalingPolicy", "ScheduledAction"}},
	"backup":       AwsService{Name: "Backup", Types: []string{"BackupPlan", "BackupSelection", "BackupVault", "RecoveryPoint"}},
	"c9":           AwsService{Name: "Cloud9", Types: []string{"EnvironmentEC2"}},
	"cfn":          AwsService{Name: "CloudFormation", Types: []string{"Stack"}},
	"cf":           AwsService{Name: "CloudFront", Types: []string{"Distribution", "StreamingDistribution"}},
	"ct":           AwsService{Name: "CloudTrail", Types: []string{"Trail"}},
	"cw":           AwsService{Name: "CloudWatch", Types: []string{"Alarm"}},
	"cb":           AwsService{Name: "CodeBuild", Types: []string{"Project"}},
	"cp":           AwsService{Name: "CodePipeline", Types: []string{"Pipeline"}},
	"config":       AwsService{Name: "Config", Types: []string{"ConfigurationRecorder", "ConformancePackCompliance", "ResourceCompliance"}},
	"datasync":     AwsService{Name: "DataSync", Types: []string{"LocationSMB"}},
	"detective":    AwsService{Name: "Detective", Types: []string{"Graph"}},
	"dynamo":       AwsService{Name: "DynamoDB", Types: []string{"Table"}},
	"ec2":          AwsService{Name: "EC2", Types: []string{"CustomerGateway", "EIP", "EgressOnlyInternetGateway", "FlowLog", "Host", "Instance", "InternetGateway", "NatGateway", "NetworkAcl", "NetworkInterface", "RegisteredHAInstance", "RouteTable", "SecurityGroup", "Subnet", "VPC", "VPCEndpoint", "VPCEndpointService", "VPCPeeringConnection", "VPNConnection", "VPNGateway", "Volume"}},
	"ecr":          AwsService{Name: "ECR", Types: []string{"PublicRepository", "RegistryPolicy", "Repository"}},
	"ecs":          AwsService{Name: "ECS", Types: []string{"Cluster", "Service", "TaskDefinition"}},
	"efs":          AwsService{Name: "EFS", Types: []string{"AccessPoint", "FileSystem"}},
	"eks":          AwsService{Name: "EKS", Types: []string{"Cluster", "FargateProfile"}},
	"emr":          AwsService{Name: "EMR", Types: []string{"SecurityConfiguration"}},
	"eb":           AwsService{Name: "ElasticBeanstalk", Types: []string{"Application", "ApplicationVersion", "Environment"}},
	"elb":          AwsService{Name: "ElasticLoadBalancing", Types: []string{"LoadBalancer"}},
	"elbv2":        AwsService{Name: "ElasticLoadBalancingV2", Types: []string{"LoadBalancer"}},
	"es":           AwsService{Name: "Elasticsearch", Types: []string{"Domain"}},
	"eventschemas": AwsService{Name: "EventSchemas", Types: []string{"Discoverer", "Registry", "RegistryPolicy"}},
	"fd":           AwsService{Name: "FraudDetector", Types: []string{"EntityType", "Label", "Outcome", "Variable"}},
	"ga":           AwsService{Name: "GlobalAccelerator", Types: []string{"Accelerator", "EndpointGroup", "Listener"}},
	"iam":          AwsService{Name: "IAM", Types: []string{"Group", "Policy", "Role", "User"}},
	"iot":          AwsService{Name: "IoT", Types: []string{"Authorizer", "Dimension", "MitigationAction", "RoleAlias", "SecurityProfile"}},
	"iota":         AwsService{Name: "IoTAnalytics", Types: []string{"Channel", "Dataset", "Datastore", "Pipeline"}},
	"iotsw":        AwsService{Name: "IoTSiteWise", Types: []string{"AssetModel", "Dashboard", "Portal", "Project"}},
	"iottw":        AwsService{Name: "IoTTwinMaker", Types: []string{"Entity"}},
	"kms":          AwsService{Name: "KMS", Types: []string{"Alias", "Key"}},
	"kav2":         AwsService{Name: "KinesisAnalyticsV2", Types: []string{"Application"}},
	"lambda":       AwsService{Name: "Lambda", Types: []string{"Alias", "Function"}},
	"ls":           AwsService{Name: "LightSail", Types: []string{"Bucket", "Certificate", "Disk", "StaticIp"}},
	"mp":           AwsService{Name: "MediaPackage", Types: []string{"PackagingGroup"}},
	"nf":           AwsService{Name: "NetworkFirewall", Types: []string{"Firewall", "FirewallPolicy", "RuleGroup"}},
	"os":           AwsService{Name: "OpenSearch", Types: []string{"Domain"}},
	"qldb":         AwsService{Name: "QLDB", Types: []string{"Ledger"}},
	"rds":          AwsService{Name: "RDS", Types: []string{"DBCluster", "DBClusterSnapshot", "DBInstance", "DBSecurityGroup", "DBSnapshot", "DBSubnetGroup", "EventSubscription", "GlobalCluster"}},
	"rs":           AwsService{Name: "Redshift", Types: []string{"Cluster", "ClusterParameterGroup", "ClusterSecurityGroup", "ClusterSnapshot", "ClusterSubnetGroup", "EventSubscription"}},
	"rh":           AwsService{Name: "ResiliencyHub", Types: []string{"ResiliencyPolicy"}},
	"r53rr":        AwsService{Name: "Route53RecoveryReadiness", Types: []string{"ReadinessCheck", "RecoveryGroup"}},
	"s3":           AwsService{Name: "S3", Types: []string{"AccountPublicAccessBlock", "Bucket", "StorageLens"}},
	"sns":          AwsService{Name: "SNS", Types: []string{"Topic"}},
	"sqs":          AwsService{Name: "SQS", Types: []string{"Queue"}},
	"ssm":          AwsService{Name: "SSM", Types: []string{"AssociationCompliance", "FileData", "ManagedInstanceInventory", "PatchCompliance"}},
	"sm":           AwsService{Name: "SecretsManager", Types: []string{"Secret"}},
	"sc":           AwsService{Name: "ServiceCatalog", Types: []string{"CloudFormationProduct", "CloudFormationProvisionedProduct", "Portfolio"}},
	"shield":       AwsService{Name: "Shield", Types: []string{"Protection"}},
	"shieldreg":    AwsService{Name: "ShieldRegional", Types: []string{"Protection"}},
	"transfer":     AwsService{Name: "Transfer", Types: []string{"Workflow"}},
	"waf":          AwsService{Name: "WAF", Types: []string{"RateBasedRule", "Rule", "RuleGroup", "WebACL"}},
	"wafreg":       AwsService{Name: "WAFRegional", Types: []string{"RateBasedRule", "Rule", "RuleGroup", "WebACL"}},
	"wafv2":        AwsService{Name: "WAFv2", Types: []string{"IPSet", "RegexPatternSet", "RuleGroup", "WebACL"}},
	"xray":         AwsService{Name: "XRay", Types: []string{"EncryptionConfig"}},
}
