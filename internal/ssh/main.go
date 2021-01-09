package ssh

import (
	flags "github.com/jessevdk/go-flags"
)

type SshCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	User    string `short:"u" long:"user" env:"AWSFUZZY_SSH_USER" default:"$USER" description:"Username to use with SSH"`
	Key     string `short:"k" long:"key" env:"AWSFUZZY_SSH_KEY" default:"~/.ssh/id_rsa" description:"Key to use with SSH"`
}

var (
	sshCommand SshCommand
)

func Init(parser *flags.Parser) {
	_, err := parser.AddCommand(
		"ssh",
		"SSH to EC2 instances",
		"SSH to EC2 instances",
		&sshCommand)

	if err != nil {
		return
	}
}
