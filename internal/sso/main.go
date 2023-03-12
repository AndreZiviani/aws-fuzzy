package sso

import (
	"github.com/AndreZiviani/aws-fuzzy/internal/afconfig"
	"github.com/AndreZiviani/aws-fuzzy/internal/awsprofile"
	"github.com/urfave/cli/v2"
)

type Login struct {
	Profile  string
	Ask      bool
	MFATOTP  string
	Verbose  bool
	Url      bool
	NoCache  bool
	profiles awsprofile.Profiles
}

type Console struct {
	Profile string
	Region  string
	Service string
	Url     bool
	Verbose bool
	NoCache bool
}

type Browser struct {
	Browser string
	Verbose bool
}

type Configure struct {
	Verbose bool
}

type CredentialProcess struct {
	Profile string
	MFATOTP string
	Verbose bool
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "sso",
		Usage: "SSO Utilities",
		Subcommands: []*cli.Command{
			{
				Name:  "login",
				Usage: "Login to AWS",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.BoolFlag{Name: "ask", Usage: "Ask before continuing"},
					&cli.StringFlag{Name: "token", Aliases: []string{"t"}, Usage: "MFA TOTP if using IAM authentication with MFA"},
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
					&cli.BoolFlag{Name: "url", Aliases: []string{"u"}, Usage: "Only print login url"},
					&cli.BoolFlag{Name: "no-cache", Aliases: []string{"n"}, Usage: "Dont use cached credentials"},
				},
				Action: func(c *cli.Context) error {
					login := NewLogin(c.String("profile"),
						c.String("token"),
						c.Bool("ask"),
						c.Bool("verbose"),
						c.Bool("url"),
						c.Bool("no-cache"),
					)

					return login.Execute(c.Context)
				},
			},
			{
				Name:  "console",
				Usage: "Open AWS Console",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "region", Aliases: []string{"r"}, Usage: "What AWS region to use", Value: "us-east-1", EnvVars: []string{"AWS_REGION", "AWS_DEFAULT_REGION"}},
					&cli.StringFlag{Name: "service", Aliases: []string{"s"}, Usage: "Open console at specific service"},
					&cli.BoolFlag{Name: "url", Aliases: []string{"u"}, Usage: "Only print login url"},
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
					&cli.BoolFlag{Name: "no-cache", Aliases: []string{"n"}, Usage: "Dont use cached credentials"},
				},
				Action: func(c *cli.Context) error {
					console := NewConsole(c.String("profile"),
						c.String("region"),
						c.String("service"),
						c.Bool("url"),
						c.Bool("verbose"),
						c.Bool("no-cache"),
					)

					return console.Execute(c.Context)
				},
			},
			{
				Name:  "browser",
				Usage: "Configure default browser",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "browser", Aliases: []string{"b"}, Usage: "Specify a default browser without prompts, e.g '-b firefox', '-b chrome'"},
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
				},
				Action: func(c *cli.Context) error {
					browser := NewBrowser(c.String("browser"),
						c.Bool("verbose"),
					)

					return browser.Execute(c.Context)
				},
			},
			{
				Name:  "configure",
				Usage: "Configure AWS SSO",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
				},
				Action: func(c *cli.Context) error {
					login := NewConfigure(c.Bool("verbose"))

					return login.Execute(c.Context)
				},
			},
			{
				Name:  "credential-process",
				Usage: "Integrate with native AWS CLI",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
					&cli.StringFlag{Name: "token", Aliases: []string{"t"}, Usage: "MFA TOTP if using IAM authentication with MFA"},
					&cli.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Enable verbose messages"},
				},
				Action: func(c *cli.Context) error {
					cp := NewCredentialProcess(c.String("profile"),
						c.String("token"),
						c.Bool("verbose"),
					)

					return cp.Execute(c.Context)
				},
			},
		},
	}

	config := afconfig.NewDefaultConfig()
	config.SetupConfigFolder()

	return &command
}
