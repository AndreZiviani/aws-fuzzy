package chart

import (
	flags "github.com/jessevdk/go-flags"
)

type Peering struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Account string `short:"a" long:"account" default:"" description:"Filter results to this account"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

type NM struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`

	// Hardcoded to us-west-2 because network manager is only available there for now
	Region string `hidden:"true" short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

type TGroutes struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Region  string `short:"r" long:"region" env:"AWS_REGION" default:"us-east-1" description:"What region to use"`
}

var (
	_peering  Peering
	_nm       NM
	_tgroutes TGroutes
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
		&_peering)

	cmd.AddCommand("nm",
		"Chart NetworkManager topology",
		"Chart NetworkManager topology",
		&_nm)

	cmd.AddCommand("tgroutes",
		"Chart TransitGateway route tables",
		"Chart TransitGateway route tables",
		&_tgroutes)
}
