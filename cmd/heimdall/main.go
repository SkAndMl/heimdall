package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/SkAndMl/heimdall/internal/scan"
)

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

func main() {
	args := os.Args
	if len(args) < 3 || args[1] != "scan" {
		exitWithUsage("expected command: scan <path>")
	}

	jsonReport := false
	explainReport := false
	maxDepth := -1
	limit := -1

	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonReport = true
		case "--explain":
			explainReport = true
		case "--max-depth":
			if i+1 >= len(args) {
				exitWithUsage("--max-depth requires a positive integer value")
			}
			depth, err := strconv.Atoi(args[i+1])
			if err != nil || depth <= 0 {
				exitWithUsage(fmt.Sprintf("invalid --max-depth value %q; expected a positive integer", args[i+1]))
			}
			maxDepth = depth
			i++
		case "--limit":
			if i+1 >= len(args) {
				exitWithUsage("--limit requires a positive integer value")
			}
			parsedLimit, err := strconv.Atoi(args[i+1])
			if err != nil || parsedLimit <= 0 {
				exitWithUsage(fmt.Sprintf("invalid --limit value %q; expected a positive integer", args[i+1]))
			}
			limit = parsedLimit
			i++
		default:
			exitWithUsage(fmt.Sprintf("unknown option %q", args[i]))
		}
	}

	scanner, err := scan.NewScanner(args[2], maxDepth)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: scan failed: %s\n", err)
		os.Exit(2)
	}

	fmt.Println(scanner.ScannerReport(limit, jsonReport, explainReport))
}
