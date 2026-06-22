package scan

import "github.com/SkAndMl/heimdall/internal/categories"

type ScanWarning struct {
	Path    string
	Type    string
	Message string
}

type Finding struct {
	Path     string
	Size     int64
	Category categories.ID
}

type Scanner struct {
	RootPath       string
	RootNode       *Node
	NumFiles       int
	NumDirectories int
	MaxDepth       int
	Warnings       []ScanWarning
	Categories     map[categories.ID][]Finding
}

type Node struct {
	Path     string
	TotSize  int64
	Type     string
	Children []*Node
	Depth    int
}

type Options struct {
	Path          string
	JSONReport    bool
	ExplainReport bool
	MaxDepth      int
	Limit         int
}
