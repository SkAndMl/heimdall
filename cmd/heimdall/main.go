package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SkAndMl/heimdall/internal/scan"
)

func main() {
	args := os.Args
	if len(args) != 3 || args[1] != "scan" {
		fmt.Println("Invalid command!")
		os.Exit(1)
	}
	absPath, err := filepath.Abs(args[2])
	if err != nil {
		fmt.Println("Cannot get path")
		os.Exit(1)
	}

	rootNode, err := scan.CreateRootNode(absPath)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(rootNode.Path)
	fmt.Println(rootNode.Depth)
	fmt.Println(float64(rootNode.TotSize) / (1024 * 1024 * 1024))

	largestFiles := rootNode.GetLargestFiles(50)
	fmt.Println("Top 50 largest files")
	for _, file := range largestFiles {
		fmt.Printf("File: %s, Size: %f\n", file.Path, float64(file.TotSize)/float64(1024*1024))
	}
}
