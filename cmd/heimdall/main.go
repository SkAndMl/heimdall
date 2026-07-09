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
	"github.com/SkAndMl/heimdall/internal/inspect"
	"github.com/SkAndMl/heimdall/internal/ps"
	"github.com/SkAndMl/heimdall/internal/run"
	"github.com/SkAndMl/heimdall/internal/stop"
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
		parsedArgs, err := cli.ParseRunArgs(args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := run.HandleRunCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	case "ps":
		parsedArgs, err := cli.ParsePsArgs(args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := ps.HandlePsCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	case "inspect":
		parsedArgs, err := cli.ParseInspectArgs(args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := inspect.HandleInspectCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	case "stop":
		parsedArgs, err := cli.ParseStopArgs(args)
		if err != nil {
			log.Fatalln(err)
		}
		if err := stop.HandleStopCommand(parsedArgs); err != nil {
			log.Fatalln(err)
		}
	default:
		log.Fatalln(fmt.Errorf("Unrecognized command\n"))
	}

}
