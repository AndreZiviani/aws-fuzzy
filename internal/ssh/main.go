package ssh

import (
	"github.com/urfave/cli/v2"
)

type Ssh struct {
	Profile string
	User    string
	Key     string
}

func Command() *cli.Command {
	command := cli.Command{
		Name:  "ssh",
		Usage: "SSH to EC2 instances",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "profile", Aliases: []string{"p"}, Usage: "What profile to use", Value: "$AWS_PROFILE", EnvVars: []string{"AWSFUZZY_PROFILE", "AWS_PROFILE"}},
			&cli.StringFlag{Name: "user", Aliases: []string{"u"}, Usage: "Username to use with SSH", Value: "$USER", EnvVars: []string{"AWSFUZZY_SSH_USER", "USER"}},
			&cli.StringFlag{Name: "key", Aliases: []string{"k"}, Usage: "Key to use with SSH", Value: "~/.ssh/id_rsa", EnvVars: []string{"AWSFUZZY_SSH_KEY"}},
		},
		Action: func(c *cli.Context) error {
			ssh := New(c.String("profile"), c.String("user"), c.String("key"))
			return ssh.Execute(c.Context)
		},
	}

	return &command
}
