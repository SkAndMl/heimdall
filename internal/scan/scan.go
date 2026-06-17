package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SkAndMl/heimdall/internal/util"
)

type Scanner struct {
	RootPath       string
	RootNode       *Node
	NumFiles       int
	NumDirectories int
	SkippedPaths   [][2]string
	MaxDepth       int
	JSONReport     bool
}

type Node struct {
	Path     string
	TotSize  int64
	Type     string
	Children []*Node
	Depth    int
}

func NewScanner(path string, maxDepth int) (*Scanner, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	scanner := Scanner{
		RootPath: absPath,
		MaxDepth: maxDepth,
	}
	rootNode, err := scanner.walkPath(scanner.RootPath, 1)
	if err != nil {
		return nil, err
	}
	scanner.RootNode = rootNode
	return &scanner, nil
}

func (s *Scanner) walkPath(path string, curDepth int) (*Node, error) {

	if s.MaxDepth != -1 && curDepth > s.MaxDepth {
		return nil, nil
	}

	if util.ShouldSkip(path) {
		return nil, nil
	}

	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}

	node := &Node{Path: path}

	if !info.IsDir() {
		node.Type = "file"
		node.TotSize = info.Size()
		node.Depth = curDepth
		s.NumFiles++
		return node, nil
	}

	node.Type = "dir"
	node.Depth = curDepth
	s.NumDirectories++

	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsPermission(err) {
			s.SkippedPaths = append(s.SkippedPaths, [2]string{path, err.Error()})
			return node, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		childNode, err := s.walkPath(filepath.Join(path, entry.Name()), curDepth+1)
		if err != nil {
			return nil, err
		}
		if childNode != nil {
			node.TotSize += childNode.TotSize
			node.Children = append(node.Children, childNode)
		}
	}
	return node, nil
}

func (s *Scanner) GetLargestEntries(nEntries int, entryType string) []*Node {
	queue := []*Node{s.RootNode}
	largestEntries := make([]*Node, 0)

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if node == nil {
			continue
		}

		if node.Type == entryType {
			largestEntries = append(largestEntries, node)
			sort.Slice(largestEntries, func(i, j int) bool {
				return largestEntries[i].TotSize >= largestEntries[j].TotSize
			})
			if len(largestEntries) >= nEntries {
				largestEntries = largestEntries[:nEntries]
			}
		}

		if node.Type == "dir" {
			queue = append(queue, node.Children...)
		}
	}

	return largestEntries
}

func (s *Scanner) ScannerReport() string {
	formatCount := func(n int) string {
		raw := fmt.Sprintf("%d", n)
		if len(raw) <= 3 {
			return raw
		}

		var b strings.Builder
		firstGroupLen := len(raw) % 3
		if firstGroupLen == 0 {
			firstGroupLen = 3
		}

		b.WriteString(raw[:firstGroupLen])
		for i := firstGroupLen; i < len(raw); i += 3 {
			b.WriteString(",")
			b.WriteString(raw[i : i+3])
		}
		return b.String()
	}

	formatSize := func(size int64) string {
		const unit = 1024
		if size < unit {
			return fmt.Sprintf("%d B", size)
		}

		value := float64(size)
		units := []string{"B", "KB", "MB", "GB", "TB", "PB"}
		unitIndex := 0
		for value >= unit && unitIndex < len(units)-1 {
			value /= unit
			unitIndex++
		}

		if value >= 100 {
			return fmt.Sprintf("%.0f %s", value, units[unitIndex])
		}
		return fmt.Sprintf("%.1f %s", value, units[unitIndex])
	}

	var b strings.Builder

	b.WriteString("Heimdall Scan Report\n\n")
	b.WriteString(fmt.Sprintf("%-21s%s\n", "Path scanned:", s.RootPath))

	totalSize := int64(0)
	if s.RootNode != nil {
		totalSize = s.RootNode.TotSize
	}

	if s.JSONReport {
		type reportEntry struct {
			Size string `json:"size"`
			Path string `json:"path"`
		}

		type reportWarning struct {
			Warning string `json:"warning"`
			Path    string `json:"path"`
		}

		report := struct {
			PathScanned        string          `json:"path_scanned"`
			TotalSize          string          `json:"total_size"`
			FilesScanned       string          `json:"files_scanned"`
			DirectoriesScanned string          `json:"directories_scanned"`
			SkippedPaths       string          `json:"skipped_paths"`
			LargestDirectories []reportEntry   `json:"largest_directories"`
			LargestFiles       []reportEntry   `json:"largest_files"`
			Warnings           []reportWarning `json:"warnings"`
		}{
			PathScanned:        s.RootPath,
			TotalSize:          formatSize(totalSize),
			FilesScanned:       formatCount(s.NumFiles),
			DirectoriesScanned: formatCount(s.NumDirectories),
			SkippedPaths:       formatCount(len(s.SkippedPaths)),
			LargestDirectories: make([]reportEntry, 0),
			LargestFiles:       make([]reportEntry, 0),
			Warnings:           make([]reportWarning, 0),
		}

		for _, dir := range s.GetLargestEntries(6, "dir") {
			if dir == nil || dir.Path == s.RootPath {
				continue
			}
			report.LargestDirectories = append(report.LargestDirectories, reportEntry{
				Size: formatSize(dir.TotSize),
				Path: dir.Path,
			})
		}

		for _, file := range s.GetLargestEntries(5, "file") {
			if file == nil {
				continue
			}
			report.LargestFiles = append(report.LargestFiles, reportEntry{
				Size: formatSize(file.TotSize),
				Path: file.Path,
			})
		}

		for _, skippedPath := range s.SkippedPaths {
			warning := skippedPath[1]
			if strings.Contains(warning, "permission denied") || strings.Contains(warning, "operation not permitted") {
				warning = "permission denied"
			}
			report.Warnings = append(report.Warnings, reportWarning{
				Warning: warning,
				Path:    skippedPath[0],
			})
		}

		reportJSON, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err.Error()
		}
		return string(reportJSON)
	}

	b.WriteString(fmt.Sprintf("%-21s%s\n", "Total size:", formatSize(totalSize)))
	b.WriteString(fmt.Sprintf("%-21s%s\n", "Files scanned:", formatCount(s.NumFiles)))
	b.WriteString(fmt.Sprintf("%-21s%s\n", "Directories scanned:", formatCount(s.NumDirectories)))
	b.WriteString(fmt.Sprintf("%-21s%s\n\n", "Skipped paths:", formatCount(len(s.SkippedPaths))))

	b.WriteString("Largest directories:\n")
	for _, dir := range s.GetLargestEntries(6, "dir") {
		if dir == nil || dir.Path == s.RootPath {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-8s %s\n", formatSize(dir.TotSize), dir.Path))
	}

	b.WriteString("\nLargest files:\n")
	for _, file := range s.GetLargestEntries(5, "file") {
		if file == nil {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-8s %s\n", formatSize(file.TotSize), file.Path))
	}

	if len(s.SkippedPaths) == 0 {
		return b.String()
	}

	b.WriteString("\nWarnings:\n")
	for _, skippedPath := range s.SkippedPaths {
		warning := skippedPath[1]
		if strings.Contains(warning, "permission denied") || strings.Contains(warning, "operation not permitted") {
			warning = "permission denied"
		}
		b.WriteString(fmt.Sprintf("  %-18s %s\n", warning+":", skippedPath[0]))
	}

	return b.String()
}
