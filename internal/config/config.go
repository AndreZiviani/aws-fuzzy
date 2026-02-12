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

func New(profile, account, region, selectFilter, filter, service, serviceType string, pager bool, limit int) *Config {
	config := Config{
		Profile: profile,
		Account: account,
		Region:  region,
		Select:  selectFilter,
		Filter:  filter,
		Service: service,
		Type:    serviceType,
		Pager:   pager,
		Limit:   limit,
	}

	return &config
}

func (p *Config) Execute(ctx context.Context) error {
	closer, err := tracing.InitTracing()
	if err != nil {
		return fmt.Errorf("failed to initialize tracing, %s", err)
	}
	defer func() { _ = closer.Close() }()

	tracer := opentracing.GlobalTracer()
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, tracer, "config")
	defer span.Finish()

	results, err := p.QueryConfig(ctx)
	if err != nil {
		return err
	}

	return p.Print(results)
}

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

func (p *Config) QueryConfig(ctx context.Context) ([]string, error) {
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
		// TODO: use local resources when no aggregators are available
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

	filter := fmt.Sprintf("resourceType like 'AWS::%s::%s'", p.Service, p.Type)
	if p.Filter != "" {
		filter += fmt.Sprintf(" and %s", p.Filter)
	}
	query := fmt.Sprintf("SELECT %s WHERE %s %s", p.Select, filter, accountFilter)

	spanQuery, _ := opentracing.StartSpanFromContext(ctx, "configquery")
	configPag := awsconfig.NewSelectAggregateResourceConfigPaginator(
		configclient,
		&awsconfig.SelectAggregateResourceConfigInput{
			ConfigurationAggregatorName: aggregators[0].ConfigurationAggregatorName,
			Expression:                  aws.String(query),
			MaxResults:                  100,
		},
	)
	results := make([]string, 0)
	for configPag.HasMorePages() {
		tmp, err := configPag.NextPage(ctx)
		if err != nil {
			fmt.Printf("failed to query config aggregator, %s\n", err)
			return nil, err
		}

		results = append(results, tmp.Results...)
	}
	spanQuery.Finish()

	span.Finish()
	return results, nil

}
