package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/SkAndMl/heimdall/internal/cli"
	"github.com/SkAndMl/heimdall/internal/config"
	"github.com/SkAndMl/heimdall/internal/ps"
	"github.com/SkAndMl/heimdall/internal/run"
)

func main() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}

	baseDir := filepath.Join(home, config.BASE_DIR)

	_, err = os.Stat(baseDir)
	if errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Join(baseDir, "sessions"), 0755); err != nil {
			log.Fatalln(err)
		}
	} else if err != nil {
		log.Fatalln(err)
	}

	args := os.Args
	if len(args) < 2 {
		log.Fatalln(fmt.Errorf("Invalid command"))
	}

	switch args[1] {
	case "run":
		parsedArgs, err := cli.ParseRunArgs(os.Args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := run.HanldeRunCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	case "ps":
		parsedArgs, err := cli.ParsePsArgs(os.Args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := ps.HandlePsCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	}

}
