package cli

import (
	"github.com/AndreZiviani/aws-fuzzy/internal/chart"
	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	"github.com/AndreZiviani/aws-fuzzy/internal/ssh"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	flags "github.com/jessevdk/go-flags"
)

var (
	Parser = flags.NewParser(nil, flags.Default)
)

func Run() {
	sso.Init(Parser)
	ssh.Init(Parser)
	config.Init(Parser)
	chart.Init(Parser)
	Parser.Parse()
}
