package chart

import (
	flags "github.com/jessevdk/go-flags"
)

type ChartCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	User    string `short:"u" long:"user" env:"AWSFUZZY_SSH_USER" default:"$USER" description:"Username to use with SSH"`
	Key     string `short:"k" long:"key" env:"AWSFUZZY_SSH_KEY" default:"~/.ssh/id_rsa" description:"Key to use with SSH"`
}

var (
	chartCommand ChartCommand
)

func Init(parser *flags.Parser) {
	cmd, err := parser.AddCommand(
		"chart",
		"Chart",
		"Chart relationship between resources",
		&struct{}{})

	if err != nil {
		return
	}

	cmd.AddCommand("peering",
		"Chart peering relationship",
		"Chart peering relationship",
		&chartCommand)
}
