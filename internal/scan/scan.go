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

type ScanWarning struct {
	Path    string
	Type    string
	Message string
}

type Scanner struct {
	RootPath       string
	RootNode       *Node
	NumFiles       int
	NumDirectories int
	Warnings       []ScanWarning
	MaxDepth       int
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
	rootNode, err := scanner.walkPath(scanner.RootPath, 0)
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

	if info.Mode()&os.ModeSymlink != 0 {
		s.Warnings = append(s.Warnings, ScanWarning{
			Path:    path,
			Type:    "symlink",
			Message: "skipped symlink to avoid double-counting",
		})
		return node, nil
	}

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
			s.Warnings = append(s.Warnings, ScanWarning{
				Path:    path,
				Type:    "permission_denied",
				Message: err.Error(),
			})
			return node, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		childNode, err := s.walkPath(childPath, curDepth+1)
		if err != nil {
			s.Warnings = append(s.Warnings, ScanWarning{
				Path:    childPath,
				Type:    "error",
				Message: err.Error(),
			})
			continue
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

func (s *Scanner) ScannerReport(limit int, jsonReport bool) string {
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

	dirLimit := 6
	fileLimit := 5
	if limit > 0 {
		dirLimit = limit
		fileLimit = limit
	}

	if jsonReport {
		type reportEntry struct {
			SizeBytes int64  `json:"size_bytes"`
			Path      string `json:"path"`
		}

		type reportWarning struct {
			Type    string `json:"type"`
			Path    string `json:"path"`
			Message string `json:"message"`
		}

		report := struct {
			PathScanned        string          `json:"path_scanned"`
			TotalSizeBytes     int64           `json:"total_size_bytes"`
			FilesScanned       int             `json:"files_scanned"`
			DirectoriesScanned int             `json:"directories_scanned"`
			SkippedPaths       int             `json:"skipped_paths"`
			LargestDirectories []reportEntry   `json:"largest_directories"`
			LargestFiles       []reportEntry   `json:"largest_files"`
			Warnings           []reportWarning `json:"warnings"`
		}{
			PathScanned:        s.RootPath,
			TotalSizeBytes:     totalSize,
			FilesScanned:       s.NumFiles,
			DirectoriesScanned: s.NumDirectories,
			SkippedPaths:       len(s.Warnings),
			LargestDirectories: make([]reportEntry, 0),
			LargestFiles:       make([]reportEntry, 0),
			Warnings:           make([]reportWarning, 0),
		}

		dirsAdded := 0
		for _, dir := range s.GetLargestEntries(dirLimit+1, "dir") {
			if dir == nil || dir.Path == s.RootPath {
				continue
			}
			report.LargestDirectories = append(report.LargestDirectories, reportEntry{
				SizeBytes: dir.TotSize,
				Path:      dir.Path,
			})
			dirsAdded++
			if dirsAdded >= dirLimit {
				break
			}
		}

		for _, file := range s.GetLargestEntries(fileLimit, "file") {
			if file == nil {
				continue
			}
			report.LargestFiles = append(report.LargestFiles, reportEntry{
				SizeBytes: file.TotSize,
				Path:      file.Path,
			})
		}

		for _, warning := range s.Warnings {
			report.Warnings = append(report.Warnings, reportWarning{
				Type:    warning.Type,
				Path:    warning.Path,
				Message: warning.Message,
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
	b.WriteString(fmt.Sprintf("%-21s%s\n\n", "Skipped paths:", formatCount(len(s.Warnings))))

	b.WriteString("Largest directories:\n")
	dirsAdded := 0
	for _, dir := range s.GetLargestEntries(dirLimit+1, "dir") {
		if dir == nil || dir.Path == s.RootPath {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-8s %s\n", formatSize(dir.TotSize), dir.Path))
		dirsAdded++
		if dirsAdded >= dirLimit {
			break
		}
	}

	b.WriteString("\nLargest files:\n")
	for _, file := range s.GetLargestEntries(fileLimit, "file") {
		if file == nil {
			continue
		}
		b.WriteString(fmt.Sprintf("  %-8s %s\n", formatSize(file.TotSize), file.Path))
	}

	if len(s.Warnings) == 0 {
		return b.String()
	}

	b.WriteString("\nWarnings:\n")
	for _, warning := range s.Warnings {
		b.WriteString(fmt.Sprintf("  %-18s %s (%s)\n", warning.Type+":", warning.Path, warning.Message))
	}

	return b.String()
}
