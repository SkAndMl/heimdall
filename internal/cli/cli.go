package cli

import (
	"fmt"

	"github.com/SkAndMl/heimdall/internal/run"
)

// heimdall run [flags] -- command

func ParseRunArgs(args []string) (*run.RunArgs, error) {
	runArgs := &run.RunArgs{}

	if len(args) < 2 || args[1] != "run" {
		return nil, fmt.Errorf("Subcommand is not run\n")
	}

	for i := 2; i < len(args); {
		if args[i] == "--" {
			runArgs.Command = args[i+1:][:]
			break
		}

		switch args[i] {
		case "--cwd", "-C":
			runArgs.Cwd = args[i+1]
			i += 2
		case "--name", "-n":
			runArgs.Name = args[i+1]
			i += 2
		case "--detach", "-d":
			runArgs.Detach = true
			i += 1
		default:
			return nil, fmt.Errorf("Unrecognized flag: %s\n", args[i])
		}
	}

	return runArgs, nil
}
