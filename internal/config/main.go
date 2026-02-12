package config

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"sort"
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
	keys := make([]string, 0, len(AwsServices))
	for k := range AwsServices {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		subcommands = append(subcommands, &cli.Command{
			Name:  k,
			Usage: fmt.Sprintf("Query %v service", AwsServices[k].Name),
			Flags: []cli.Flag{
				&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
				&cli.BoolFlag{Name: "pager", Usage: "Pipe output to less", Value: false},
				&cli.StringFlag{Name: "account", Aliases: []string{"a"}, Usage: "Filter Config resources to this account"},
				&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
				&cli.StringFlag{Name: "select", Aliases: []string{"s"}, Usage: "Custom select to filter results", Value: "resourceId, accountId, awsRegion, configuration, tags"},
				&cli.StringFlag{Name: "filter", Aliases: []string{"f"}, Usage: "Complete custom query"},
				&cli.IntFlag{Name: "limit", Aliases: []string{"l"}, Usage: "Limit the number of results", Value: 0},
				&cli.StringFlag{Name: "type", Aliases: []string{"t"}, Value: "%", Usage: fmt.Sprintf("Filter results to only one of the following types: %q", AwsServices[k].Types)},
			},
			Action: func(c *cli.Context) error {
				serviceType := c.String("type")
				if serviceType != "%" {
					if service, ok := AwsServices[c.Command.Name]; ok {
						if !contains(service.Types, serviceType) {
							return fmt.Errorf("could not find type '%s' for service '%s'", serviceType, service.Name)
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

//go:generate go run generate.go
