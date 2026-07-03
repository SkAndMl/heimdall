package main

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/SkAndMl/heimdall/internal/cli"
	"github.com/SkAndMl/heimdall/internal/config"
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

	parsedArgs, err := cli.ParseRunArgs(os.Args)
	if err != nil {
		log.Fatalln(err)
	}
	if err := run.Run(parsedArgs); err != nil {
		log.Fatal(err)
	}
}
