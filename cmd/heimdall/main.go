package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/SkAndMl/heimdall/internal/scan"
)

func main() {
	args := os.Args
	if len(args) < 3 || args[1] != "scan" {
		fmt.Println("Invalid command!")
		os.Exit(1)
	}

	jsonReport := false
	maxDepth := -1

	for i := 3; i < len(args); i++ {
		switch args[i] {
		case "--json":
			jsonReport = true
		case "--max-depth":
			if i+1 >= len(args) {
				fmt.Println("Missing value for --max-depth")
				os.Exit(1)
			}
			depth, err := strconv.Atoi(args[i+1])
			if err != nil {
				fmt.Println("Invalid value for --max-depth")
				os.Exit(1)
			}
			maxDepth = depth
			i++
		default:
			fmt.Println("Invalid command!")
			os.Exit(1)
		}
	}

	scanner, err := scan.NewScanner(args[2], maxDepth)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	scanner.JSONReport = jsonReport

	fmt.Println(scanner.ScannerReport())
}
