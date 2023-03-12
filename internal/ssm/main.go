package ssm

import (
	"github.com/urfave/cli/v2"
)

type Session struct {
	Profile string
	Region  string
	Shell   string
}

type PortForward struct {
	Profile string
	Region  string
	Ports   string
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "ssm",
		Usage: "Interact with EC2 instances via SSM",
		Subcommands: []*cli.Command{
			{
				Name:  "session",
				Usage: "Start a session on a EC2 instance",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
					&cli.StringFlag{Name: "shell", Aliases: []string{"s"}, Value: "bash", Usage: "What shell to use on the remote instance"},
				},
				Action: func(c *cli.Context) error {
					session := NewSession(c.String("profile"),
						c.String("region"),
						c.String("shell"),
					)

					return session.Execute(c.Context)
				},
			},
			{
				Name:  "portforward",
				Usage: "Start a portforwarding session on a EC2 instance",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
					&cli.StringFlag{Name: "ports", Value: "8080:localhost:80", Usage: "Binds remote port to local, '<local port>:<remote host>:<remote port>'"},
				},
				Action: func(c *cli.Context) error {
					pf := NewPortForward(c.String("profile"),
						c.String("region"),
						c.String("ports"),
					)

					return pf.Execute(c.Context)
				},
			},
		},
	}

	return &command
}
