package eks

import (
	"github.com/urfave/cli/v2"
)

type GetToken struct {
	Profile     string
	ClusterName string
	Region      string
	NoCache     bool
	Verbose     bool
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "eks",
		Usage: "EKS Utilities",
		Subcommands: []*cli.Command{
			{
				Name:  "get-token",
				Usage: "Get an authentication token for an EKS cluster",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "cluster-name", Aliases: []string{"c"}, Usage: "The name of the EKS cluster", Required: true},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
					&cli.BoolFlag{Name: "no-cache", Aliases: []string{"n"}, Usage: "Don't use cached EKS token"},
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
				},
				Action: func(c *cli.Context) error {
					gt := NewGetToken(
						c.String("profile"),
						c.String("cluster-name"),
						c.String("region"),
						c.Bool("no-cache"),
						c.Bool("verbose"),
					)
					return gt.Execute(c.Context)
				},
			},
		},
	}

	return &command
}
