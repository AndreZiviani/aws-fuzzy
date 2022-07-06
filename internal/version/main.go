package version

import (
	flags "github.com/jessevdk/go-flags"
)

type Version struct {
}

var (
	versionTag string
	version    Version
)

func Init(parser *flags.Parser) {
	_, err := parser.AddCommand(
		"version",
		"Show aws-fuzzy version",
		"Show aws-fuzzy version",
		&version)

	if err != nil {
		return
	}
}
