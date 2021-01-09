package sso

import (
	flags "github.com/jessevdk/go-flags"
)

type LoginCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Ask     bool   `long:"ask" env:"AWSFUZZY_ASK" description:"Ask before continuing"`
}

type ConfigureCommand struct{}

var (
	loginCommand     LoginCommand
	configureCommand ConfigureCommand
)

func Init(parser *flags.Parser) {
	cmd, err := parser.AddCommand(
		"sso",
		"SSO Utilities",
		"Utilities developed to ease operation and configuration of AWS SSO",
		&struct{}{})

	if err != nil {
		return
	}

	cmd.AddCommand("login",
		"Login to AWS SSO",
		"Login to AWS SSO",
		&loginCommand)

	cmd.AddCommand("configure",
		"Configure AWS SSO",
		"Configure local profiles with AWS accounts available from SSO",
		&configureCommand)
}
