package chart

import (
	flags "github.com/jessevdk/go-flags"
)

type PeeringCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Account string `short:"a" long:"account" default:"" description:"Filter results to this account"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

type NMCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`

	// Hardcoded to us-west-2 because network manager is only available there for now
	Region string `hidden:"true" short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

type TGroutesCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

var (
	peeringCommand  PeeringCommand
	nmCommand       NMCommand
	tgroutesCommand TGroutesCommand
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
		&peeringCommand)

	cmd.AddCommand("nm",
		"Chart NetworkManager topology",
		"Chart NetworkManager topology",
		&nmCommand)

	cmd.AddCommand("tgroutes",
		"Chart TransitGateway route tables",
		"Chart TransitGateway route tables",
		&tgroutesCommand)
}
