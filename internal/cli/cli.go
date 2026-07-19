package cli

import (
	"fmt"
	"strconv"

	"github.com/SkAndMl/heimdall/internal/inspect"
	"github.com/SkAndMl/heimdall/internal/logs"
	"github.com/SkAndMl/heimdall/internal/ps"
	"github.com/SkAndMl/heimdall/internal/run"
	"github.com/SkAndMl/heimdall/internal/session"
	"github.com/SkAndMl/heimdall/internal/stop"
)

// heimdall run [flags] -- command

func ParseRunArgs(args []string) (*run.RunArgs, error) {
	runArgs := &run.RunArgs{}

	if len(args) < 2 || (args[1] != "run" && args[1] != "_run-supervisor") {
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

	if len(runArgs.Command) == 0 {
		return nil, fmt.Errorf("No command available\n")
	}

	return runArgs, nil
}

// heimdall ps [flags]
// --all, --status status_name, --json

func ParsePsArgs(args []string) (*ps.PsArgs, error) {
	psArgs := &ps.PsArgs{}

	if len(args) < 2 || args[1] != "ps" {
		return nil, fmt.Errorf("Invalid command format\n")
	}

	for i := 2; i < len(args); {
		switch args[i] {
		case "--all":
			psArgs.All = true
			i += 1
		case "--status":
			if i+1 == len(args) {
				return nil, fmt.Errorf("No value passed for '--status' flag\n")
			}
			psArgs.Status = session.Status(args[i+1])
			i += 2
		case "--json":
			psArgs.JSONOutput = true
			i += 1
		default:
			return nil, fmt.Errorf("Invalid command format\n")
		}
	}

	return psArgs, nil
}

// heimdall inspect <session-ref>
func ParseInspectArgs(args []string) (*inspect.InspectArgs, error) {
	if len(args) != 3 || args[1] != "inspect" {
		return nil, fmt.Errorf("Invalid command\n")
	}

	inspectArgs := &inspect.InspectArgs{
		SessionRef: args[2],
	}
	return inspectArgs, nil
}

func ParseStopArgs(args []string) (*stop.StopArgs, error) {
	if len(args) != 3 || args[1] != "stop" {
		return nil, fmt.Errorf("Invalid command\n")
	}

	return &stop.StopArgs{SessionRef: args[2]}, nil
}

func ParseLogsArgs(args []string) (*logs.LogArgs, error) {
	if len(args) < 3 || args[1] != "logs" {
		return nil, fmt.Errorf("Invalid command\n")
	}

	logArgs := &logs.LogArgs{
		SessionRef: args[2],
	}

	for i := 3; i < len(args); {
		switch args[i] {
		case "--stderr":
			logArgs.StderrFlag = true
			i += 1
		case "--follow":
			logArgs.FollowFlag = true
			i += 1
		case "--tail":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("--tail option requires an integer value\n")
			}
			lastNLines, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, err
			}
			if lastNLines <= 0 {
				return nil, fmt.Errorf("Value for --tail option should be +ve\n")
			}
			logArgs.LastNLines = lastNLines
			i += 2
		default:
			return nil, fmt.Errorf("Unrecognized option %q\n", args[i])
		}
	}
	return logArgs, nil
}
