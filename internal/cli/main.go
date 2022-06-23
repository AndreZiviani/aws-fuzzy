package cli

import (
	"github.com/AndreZiviani/aws-fuzzy/internal/chart"
	"github.com/AndreZiviani/aws-fuzzy/internal/config"
	"github.com/AndreZiviani/aws-fuzzy/internal/ssh"
	"github.com/AndreZiviani/aws-fuzzy/internal/ssm"
	"github.com/AndreZiviani/aws-fuzzy/internal/sso"
	flags "github.com/jessevdk/go-flags"
)

var (
	Parser = flags.NewParser(nil, flags.Default)
)

func Run() {
	ssh.Init(Parser)
	config.Init(Parser)
	chart.Init(Parser)
	sso.Init(Parser)
	ssm.Init(Parser)
	Parser.Parse()
}
