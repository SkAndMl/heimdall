package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/SkAndMl/heimdall/internal/clean"
	"github.com/SkAndMl/heimdall/internal/scan"
)

type CLIArgs struct {
	Subcommand string
	ScanArgs   scan.Options
	CleanArgs  clean.Options
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

func exitWithUsage(message string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n\n%s\n", message, usage())
	os.Exit(1)
}

func parseScanArgs(args []string) scan.Options {
	scanArgs := scan.Options{
		JSONReport:    false,
		ExplainReport: false,
		MaxDepth:      -1,
		Limit:         -1,
	}

	if len(args) < 3 {
		exitWithUsage("expected command: scan <path>")
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
				exitWithUsage("--max-depth requires a positive integer value")
			}
			depth, err := strconv.Atoi(args[i+1])
			if err != nil || depth <= 0 {
				exitWithUsage(fmt.Sprintf("invalid --max-depth value %q; expected a positive integer", args[i+1]))
			}
			scanArgs.MaxDepth = depth
			i++
		case "--limit":
			if i+1 >= len(args) {
				exitWithUsage("--limit requires a positive integer value")
			}
			parsedLimit, err := strconv.Atoi(args[i+1])
			if err != nil || parsedLimit <= 0 {
				exitWithUsage(fmt.Sprintf("invalid --limit value %q; expected a positive integer", args[i+1]))
			}
			scanArgs.Limit = parsedLimit
			i++
		default:
			exitWithUsage(fmt.Sprintf("unknown option %q", args[i]))
		}
	}

	return scanArgs
}

func parseCleanArgs(args []string) clean.Options {
	cleanArgs := clean.Options{
		DryRun:      false,
		Interactive: false,
	}

	if len(args) != 4 {
		exitWithUsage("clean subcommand is of invalid format")
	}

	cleanArgs.Path = args[2]
	switch args[3] {
	case "--dry-run":
		cleanArgs.DryRun = true
	case "--interactive":
		cleanArgs.Interactive = true
	default:
		exitWithUsage("invalid option")
	}

	return cleanArgs
}

func ParseArgs(args []string) *CLIArgs {

	if len(args) <= 2 {
		exitWithUsage("command is of invalid format")
	}

	cliArgs := &CLIArgs{Subcommand: args[1]}

	switch args[1] {
	case "scan":
		cliArgs.ScanArgs = parseScanArgs(args)
	case "clean":
		cliArgs.CleanArgs = parseCleanArgs(args)
	default:
		exitWithUsage("unrecognized subcommand")
	}

	return cliArgs
}

func Run(args []string) {
	parsedArgs := ParseArgs(args)
	switch parsedArgs.Subcommand {
	case "scan":
		scanner, err := scan.NewScanner(parsedArgs.ScanArgs.Path, parsedArgs.ScanArgs.MaxDepth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: scan failed: %s\n", err)
			os.Exit(2)
		}

		fmt.Println(scanner.ScannerReport(parsedArgs.ScanArgs.Limit, parsedArgs.ScanArgs.JSONReport, parsedArgs.ScanArgs.ExplainReport))
	case "clean":
		clean.Clean(parsedArgs.CleanArgs)
	default:
		exitWithUsage("unrecognized subcommand")
	}
}
