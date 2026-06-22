package main

import (
	"os"

	"github.com/SkAndMl/heimdall/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args, os.Stdout, os.Stderr))
}
