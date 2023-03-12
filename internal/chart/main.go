package chart

import (
	"github.com/urfave/cli/v2"
)

type Peering struct {
	Profile string
	Account string
	Region  string
}

type NM struct {
	Profile string
}

type TGRoutes struct {
	Profile string
	Region  string
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "chart",
		Usage: "Chart relationship between resources",
		Subcommands: []*cli.Command{
			{
				Name:  "peering",
				Usage: "Chart VPC peering relationship",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "account", Aliases: []string{"a"}, Usage: "Filter Config resources to this account"},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
				},
				Action: func(c *cli.Context) error {
					peering := NewPeering(c.String("profile"),
						c.String("account"),
						c.String("region"),
					)

					return peering.Execute(c.Context)
				},
			},
			{
				Name:  "nm",
				Usage: "Chart NetworkManager topology",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
				},
				Action: func(c *cli.Context) error {
					nm := NewNM(c.String("profile"))

					return nm.Execute(c.Context)
				},
			},
			{
				Name:  "tgroutes",
				Usage: "Chart TransitGateway route tables",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
				},
				Action: func(c *cli.Context) error {
					tgroutes := NewTGRoutes(c.String("profile"),
						c.String("region"),
					)

					return tgroutes.Execute(c.Context)
				},
			},
		},
	}

	return &command
}
