package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/SkAndMl/heimdall/internal/clean"
	"github.com/SkAndMl/heimdall/internal/scan"
)

type CLIArgs struct {
	Subcommand string
	ScanArgs   scan.Options
	CleanArgs  clean.Options
}

type UsageError struct {
	Message string
}

func (e UsageError) Error() string {
	return e.Message
}

func usage() string {
	return `Usage:
  heimdall scan <path> [--json] [--explain] [--max-depth <depth>] [--limit <count>]

Examples:
  heimdall scan ~
  heimdall scan ~ --json
  heimdall scan ~ --explain
  heimdall scan ~ --max-depth 2
  heimdall scan ~ --limit 25
  heimdall scan ~ --json --explain --max-depth 2 --limit 25`
}

func parseScanArgs(args []string) (scan.Options, error) {
	scanArgs := scan.Options{
		JSONReport:    false,
		ExplainReport: false,
		MaxDepth:      -1,
		Limit:         -1,
	}

	if len(args) < 3 {
		return scanArgs, UsageError{Message: "expected command: scan <path>"}
	}

	scanArgs.Path = args[2]
	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--json":
			scanArgs.JSONReport = true
		case "--explain":
			scanArgs.ExplainReport = true
		case "--max-depth":
			if i+1 >= len(args) {
				return scanArgs, UsageError{Message: "--max-depth requires a positive integer value"}
			}
			depth, err := strconv.Atoi(args[i+1])
			if err != nil || depth <= 0 {
				return scanArgs, UsageError{Message: fmt.Sprintf("invalid --max-depth value %q; expected a positive integer", args[i+1])}
			}
			scanArgs.MaxDepth = depth
			i++
		case "--limit":
			if i+1 >= len(args) {
				return scanArgs, UsageError{Message: "--limit requires a positive integer value"}
			}
			parsedLimit, err := strconv.Atoi(args[i+1])
			if err != nil || parsedLimit <= 0 {
				return scanArgs, UsageError{Message: fmt.Sprintf("invalid --limit value %q; expected a positive integer", args[i+1])}
			}
			scanArgs.Limit = parsedLimit
			i++
		default:
			return scanArgs, UsageError{Message: fmt.Sprintf("unknown option %q", args[i])}
		}
	}

	return scanArgs, nil
}

func parseCleanArgs(args []string) (clean.Options, error) {
	cleanArgs := clean.Options{
		DryRun:      false,
		Interactive: false,
	}

	if len(args) != 4 {
		return cleanArgs, UsageError{Message: "clean subcommand is of invalid format"}
	}

	cleanArgs.Path = args[2]
	switch args[3] {
	case "--dry-run":
		cleanArgs.DryRun = true
	case "--interactive":
		cleanArgs.Interactive = true
	default:
		return cleanArgs, UsageError{Message: "invalid option"}
	}

	return cleanArgs, nil
}

func ParseArgs(args []string) (*CLIArgs, error) {

	if len(args) < 2 {
		return nil, UsageError{Message: "command is of invalid format"}
	}

	cliArgs := &CLIArgs{Subcommand: args[1]}

	switch args[1] {
	case "scan":
		scanArgs, err := parseScanArgs(args)
		if err != nil {
			return nil, err
		}
		cliArgs.ScanArgs = scanArgs
	case "clean":
		cleanArgs, err := parseCleanArgs(args)
		if err != nil {
			return nil, err
		}
		cliArgs.CleanArgs = cleanArgs
	default:
		return nil, UsageError{Message: "unrecognized subcommand"}
	}

	return cliArgs, nil
}

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := run(args, stdout); err != nil {
		var usageErr UsageError
		if errors.As(err, &usageErr) {
			fmt.Fprintf(stderr, "Error: %s\n\n%s\n", usageErr.Message, usage())
			return 1
		}

		fmt.Fprintf(stderr, "Error: %s\n", err)
		return 2
	}

	return 0
}

func run(args []string, stdout io.Writer) error {
	parsedArgs, err := ParseArgs(args)
	if err != nil {
		return err
	}

	switch parsedArgs.Subcommand {
	case "scan":
		scanner, err := scan.NewScanner(parsedArgs.ScanArgs.Path, parsedArgs.ScanArgs.MaxDepth)
		if err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}

		fmt.Fprintln(stdout, scanner.ScannerReport(parsedArgs.ScanArgs.Limit, parsedArgs.ScanArgs.JSONReport, parsedArgs.ScanArgs.ExplainReport))
	case "clean":
		report, err := clean.Clean(parsedArgs.CleanArgs)
		if err != nil {
			return fmt.Errorf("clean failed: %w", err)
		}
		if report != "" {
			fmt.Fprintln(stdout, report)
		}
	default:
		return UsageError{Message: "unrecognized subcommand"}
	}

	return nil
}
