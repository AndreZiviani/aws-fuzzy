package chart

import (
	flags "github.com/jessevdk/go-flags"
)

type ChartCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

type NMCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

var (
	chartCommand ChartCommand
	nmCommand    NMCommand
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

	cmd.AddCommand("nm",
		"Chart NetworkManager topology",
		"Chart NetworkManager topology",
		&nmCommand)
}
