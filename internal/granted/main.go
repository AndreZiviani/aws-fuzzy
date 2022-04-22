package granted

import (
	grantedconfig "github.com/common-fate/granted/pkg/config"
	flags "github.com/jessevdk/go-flags"
)

type LoginCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Ask     bool   `long:"ask" env:"AWSFUZZY_ASK" description:"Ask before continuing"`
}

type ConsoleCommand struct {
	Profile string `short:"p" long:"profile" env:"AWS_PROFILE" default:"default" description:"What profile to use"`
	Url     bool   `short:"u" long:"url" description:"Only print login url"`
}

type BrowserCommand struct {
	Browser string `short:"b" long:"browser" description:"Specify a default browser without prompts, e.g '-b firefox', '-b chrome'"`
}

var (
	loginCommand   LoginCommand
	consoleCommand ConsoleCommand
	browserCommand BrowserCommand
)

func Init(parser *flags.Parser) {
	grantedconfig.SetupConfigFolder()

	cmd, err := parser.AddCommand(
		"granted",
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
		&loginCommand)

	cmd.AddCommand("console",
		"Open AWS Console",
		"Open AWS Console",
		&consoleCommand)

	cmd.AddCommand("browser",
		"Configure default browser",
		"Configure default browser, configuration is stored in ~/.dgranted",
		&browserCommand)

}
