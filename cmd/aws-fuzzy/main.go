package main

import (
	"github.com/AndreZiviani/aws-fuzzy/internal/cli"
	"os"
)

func main() {
	if cli.Run() != nil {
		os.Exit(1)
	}
}
