package config

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	"github.com/AndreZiviani/aws-fuzzy/internal/tracing"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awsconfig "github.com/aws/aws-sdk-go-v2/service/configservice"
	opentracing "github.com/opentracing/opentracing-go"
	//"github.com/opentracing/opentracing-go/log"
)

// Pretty print json output
func (p *Config) Print(slices []string) error {

	var prettyJSON bytes.Buffer

	tmp := strings.Join(slices[:], ",")
	tmp = fmt.Sprintf("[%s]", tmp)
	_ = json.Indent(&prettyJSON, []byte(tmp), "", "  ")

	if p.Pager {
		// less
		cmd := exec.Command("less")
		cmd.Stdin = strings.NewReader(prettyJSON.String())
		cmd.Stdout = os.Stdout
		return cmd.Run()
	}

	fmt.Printf("%s\n", prettyJSON.String())
	return nil
}

func (p *Config) QueryConfig(ctx context.Context, subservice string) ([]string, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "config")

	login := sso.Login{Profile: p.Profile}
	creds, err := login.GetCredentials(ctx)
	if err != nil {
		return nil, err
	}

	cfg, err := sso.NewAwsConfig(ctx, creds, config.WithRegion(p.Region))
	if err != nil {
		return nil, err
	}

	configclient := awsconfig.NewFromConfig(cfg)

	// Searching for available aggregators
	spanGetAggregators, tmpctx := opentracing.StartSpanFromContext(ctx, "configgetaggregators")
	tmp, err := configclient.DescribeConfigurationAggregators(tmpctx,
		&awsconfig.DescribeConfigurationAggregatorsInput{},
	)
	if err != nil {
		fmt.Printf("failed to describe configuration aggregators, %s\n", err)
		return nil, err
	}

	aggregators := tmp.ConfigurationAggregators

	if len(aggregators) == 0 {
		return nil, errors.New("could not find any aggregators")
	}

	spanGetAggregators.Finish()

	// Filter results to an account, if specified by the user
	accountFilter := ""
	if p.Account != "" {
		account, err := login.GetProfile(p.Account)
		if account == nil {
			fmt.Printf("failed to get account %s, %s\n", p.Account, err)
			return nil, err
		}
		accountFilter = fmt.Sprintf(" AND accountId like '%s'", account.AWSConfig.SSOAccountID)
	}

	spanQuery, tmpctx := opentracing.StartSpanFromContext(ctx, "configquery")

	filter := fmt.Sprintf("resourceType like 'AWS::%s::%s'", p.Service, subservice)
	if p.Filter != "" {
		filter = p.Filter
	}
	query := fmt.Sprintf("SELECT %s WHERE %s %s", p.Select, filter, accountFilter)

	result, err := configclient.SelectAggregateResourceConfig(tmpctx,
		&awsconfig.SelectAggregateResourceConfigInput{
			ConfigurationAggregatorName: aggregators[0].ConfigurationAggregatorName,
			Expression:                  aws.String(query),
			MaxResults:                  100,
		},
	)
	if err != nil {
		fmt.Printf("failed to query config aggregator, %s\n", err)
		return nil, err
	}
	spanQuery.Finish()

	span.Finish()
	return result.Results, nil

}

func (p *Config) wrapper(args []string, subservice string) error {
	ctx := context.Background()

	closer, err := tracing.InitTracing()
	if err != nil {
		fmt.Printf("failed to initialize tracing, %s\n", err)
	}
	defer closer.Close()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "config")
	defer span.Finish()

	results, err := p.QueryConfig(ctx, subservice)
	if err != nil {
		return err
	}

	return p.Print(results)

}

func (p *Ec2Config) Execute(args []string) error {
	tmp := Config{
		Profile: p.Profile,
		Pager:   p.Pager,
		Account: p.Account,
		Region:  p.Region,
		Select:  p.Select,
		Filter:  p.Filter,
		Limit:   p.Limit,
		Service: p.Service,
	}
	return tmp.wrapper(args, p.Type)

}

func (p *Config) Execute(args []string) error {
	return p.wrapper(args, "%")
}
