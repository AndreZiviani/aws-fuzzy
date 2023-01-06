package ssm

import (
	flags "github.com/jessevdk/go-flags"
)

type Session struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" description:"What region to use, if not specified defaults to $AWS_DEFAULT_REGION or us-east-1"`
	Shell   string `short:"s" long:"shell" default:"bash" description:"What shell to use on the remote instance"`
}

type PortForward struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" description:"What region to use, if not specified defaults to $AWS_DEFAULT_REGION or us-east-1"`
	Ports   string `long:"ports" default:"8080:localhost:80" description:"Binds remote port to local, '<local>:<remote host>:<remote>'"`
}

var (
	session     Session
	portforward PortForward
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

	cmd.AddCommand(
		"portforward",
		"Start a portforwarding session on a EC2 instance",
		"Start a portforwarding session on a EC2 instance",
		&portforward)

}
