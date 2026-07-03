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
			if i+1 >= len(args) {
				return nil, fmt.Errorf("Missing command after --\n")
			}
			runArgs.Command = args[i+1:][:]
			break
		}

		switch args[i] {
		case "--cwd", "-C":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("Missing value for --cwd\n")
			}
			runArgs.Cwd = args[i+1]
			i += 2
		case "--name", "-n":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("Missing value for --name\n")
			}
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
