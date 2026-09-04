package ssm

import (
	"github.com/urfave/cli/v2"
)

type Session struct {
	Profile        string
	Region         string
	Shell          string
	Instance       string
	NonInteractive bool
}

type PortForward struct {
	Profile        string
	Region         string
	Ports          string
	Instance       string
	NonInteractive bool
}

type Run struct {
	Profile        string
	Region         string
	Instance       string
	Command        string
	Timeout        int
	NonInteractive bool
}

// flags shared by every ssm subcommand
func commonFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
		&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
		&cli.StringFlag{Name: "instance", Aliases: []string{"i"}, Usage: "Target instance, by instance id, Name tag or private IP (omit to pick interactively)"},
		&cli.BoolFlag{Name: "non-interactive", Usage: "Never open the instance picker, fail instead if the target is missing or ambiguous"},
	}
}

func withFlags(extra ...cli.Flag) []cli.Flag {
	return append(commonFlags(), extra...)
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "ssm",
		Usage: "Interact with EC2 instances via SSM",
		Subcommands: []*cli.Command{
			{
				Name:  "session",
				Usage: "Start a session on a EC2 instance",
				Flags: withFlags(
					&cli.StringFlag{Name: "shell", Aliases: []string{"s"}, Value: "bash", Usage: "What shell to use on the remote instance"},
				),
				Action: func(c *cli.Context) error {
					session := NewSession(c.String("profile"),
						c.String("region"),
						c.String("shell"),
						c.String("instance"),
						c.Bool("non-interactive"),
					)

					return session.Execute(c.Context)
				},
			},
			{
				Name:  "portforward",
				Usage: "Start a portforwarding session on a EC2 instance",
				Flags: withFlags(
					&cli.StringFlag{Name: "ports", Value: "8080:localhost:80", Usage: "Binds remote port to local, '<local port>:<remote host>:<remote port>'"},
				),
				Action: func(c *cli.Context) error {
					pf := NewPortForward(c.String("profile"),
						c.String("region"),
						c.String("ports"),
						c.String("instance"),
						c.Bool("non-interactive"),
					)

					return pf.Execute(c.Context)
				},
			},
			{
				Name:  "run",
				Usage: "Run a command on a EC2 instance and print its output",
				Flags: withFlags(
					&cli.StringFlag{Name: "command", Aliases: []string{"c"}, Required: true, Usage: "Command to run on the remote instance"},
					&cli.IntFlag{Name: "timeout", Value: 60, Usage: "How long to wait for the command to finish, in seconds"},
				),
				Action: func(c *cli.Context) error {
					run := NewRun(c.String("profile"),
						c.String("region"),
						c.String("instance"),
						c.String("command"),
						c.Int("timeout"),
						c.Bool("non-interactive"),
					)

					return run.Execute(c.Context)
				},
			},
		},
	}

	return &command
}
