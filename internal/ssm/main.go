package ssm

import (
	flags "github.com/jessevdk/go-flags"
)

type Session struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" description:"What region to use, if not specified defaults to $AWS_DEFAULT_REGION or us-east-1"`
}

var (
	session Session
)

func Init(parser *flags.Parser) {
	cmd, err := parser.AddCommand(
		"ssm",
		"Interact with EC2 instances via SSM",
		"Interact with EC2 instances via SSM",
		&struct{}{})

	if err != nil {
		return
	}

	cmd.AddCommand(
		"session",
		"Start a session on a EC2 instance",
		"Start a session on a EC2 instance",
		&session)

	if err != nil {
		return
	}
}
