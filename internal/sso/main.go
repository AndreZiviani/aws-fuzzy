package sso

import (
	"github.com/common-fate/granted/pkg/cfaws"
	grantedconfig "github.com/common-fate/granted/pkg/config"
	flags "github.com/jessevdk/go-flags"
)

type Login struct {
	Profile  string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Ask      bool   `long:"ask" env:"AWSFUZZY_ASK" description:"Ask before continuing"`
	MFATOTP  string `short:"t" long:"token" description:"MFA TOTP if using IAM authentication with MFA"`
	Verbose  bool   `short:"v" long:"verbose" description:"Enable verbose messages"`
	profiles cfaws.CFSharedConfigs
}

type Console struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Url     bool   `short:"u" long:"url" description:"Only print login url"`
	Verbose bool   `short:"v" long:"verbose" description:"Enable verbose messages"`
}

type Browser struct {
	Browser string `short:"b" long:"browser" description:"Specify a default browser without prompts, e.g '-b firefox', '-b chrome'"`
	Verbose bool   `short:"v" long:"verbose" description:"Enable verbose messages"`
}

type Configure struct {
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose messages"`
}

var (
	login     Login
	console   Console
	browser   Browser
	configure Configure
)

func Init(parser *flags.Parser) {
	grantedconfig.SetupConfigFolder()

	cmd, err := parser.AddCommand(
		"sso",
		"SSO Utilities",
		"Utilities developed to ease operation and configuration of AWS SSO.\n"+
			"This is mostly imported from common-fate/granted so some log messages may display 'granted' as the application name",
		&struct{}{})

	if err != nil {
		return
	}

	cmd.AddCommand("login",
		"Login to AWS",
		"Login to AWS",
		&login)

	cmd.AddCommand("console",
		"Open AWS Console",
		"Open AWS Console",
		&console)

	cmd.AddCommand("browser",
		"Configure default browser",
		"Configure default browser, configuration is stored in ~/.dgranted",
		&browser)

	cmd.AddCommand("configure",
		"Configure AWS SSO",
		"Configure local profiles with AWS accounts available from SSO",
		&configure)

}
