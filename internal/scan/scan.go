package scan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SkAndMl/heimdall/internal/categories"
	"github.com/SkAndMl/heimdall/internal/detectors"
	"github.com/SkAndMl/heimdall/internal/presentation"
	"github.com/SkAndMl/heimdall/internal/util"
)

func NewScanner(path string, maxDepth int) (*Scanner, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	scanner := Scanner{
		RootPath:   absPath,
		MaxDepth:   maxDepth,
		Categories: make(map[categories.ID][]Finding),
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

		fileCategory := categories.ClassifyFile(path)
		if fileCategory != categories.Unknown {
			s.Categories[fileCategory] = append(s.Categories[fileCategory], Finding{
				Path:     path,
				Size:     info.Size(),
				Category: fileCategory,
			})
		}

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

	dirType := detectors.ClassifyDir(path)
	if dirType != categories.Unknown {
		s.Categories[dirType] = append(s.Categories[dirType], Finding{
			Path:     path,
			Size:     node.TotSize,
			Category: dirType,
		})
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

func (s *Scanner) ScannerReport(limit int, jsonReport bool, explainReport bool) string {
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

	if explainReport {
		type explanationSummary struct {
			Label     string `json:"label"`
			Group     string `json:"group"`
			SizeBytes int64  `json:"size_bytes"`
			Risk      string `json:"risk"`
			Why       string `json:"why"`
			Action    string `json:"action"`
		}

		summariesByGroup := make(map[string]explanationSummary)
		for category, findings := range s.Categories {
			explanation, ok := categories.Lookup(category)
			if !ok {
				continue
			}

			summary := summariesByGroup[explanation.Group]
			summary.Label = explanation.Label
			summary.Group = explanation.Group
			summary.Risk = explanation.Risk
			summary.Why = explanation.Why
			summary.Action = explanation.Action

			for _, finding := range findings {
				summary.SizeBytes += finding.Size
			}
			summariesByGroup[explanation.Group] = summary
		}

		summaries := make([]explanationSummary, 0, len(summariesByGroup))
		for _, summary := range summariesByGroup {
			summaries = append(summaries, summary)
		}
		sort.Slice(summaries, func(i, j int) bool {
			return summaries[i].SizeBytes > summaries[j].SizeBytes
		})

		if limit > 0 && len(summaries) > limit {
			summaries = summaries[:limit]
		}

		if jsonReport {
			reportJSON, err := json.MarshalIndent(summaries, "", "  ")
			if err != nil {
				return err.Error()
			}
			return string(reportJSON)
		}

		if len(summaries) == 0 {
			return "No cleanup categories found yet."
		}

		var explainBuilder strings.Builder
		explainBuilder.WriteString(presentation.Brand("◉ HEIMDALL"))
		explainBuilder.WriteString("\n\n")
		explainBuilder.WriteString(presentation.Title("Cleanup categories"))
		explainBuilder.WriteString("\n")
		explainBuilder.WriteString(presentation.Divider("──────────────────────────────────────────────────────────────────"))
		explainBuilder.WriteString("\n\n")
		for i, summary := range summaries {
			if i > 0 {
				explainBuilder.WriteString("\n")
			}
			explainBuilder.WriteString(fmt.Sprintf("%s   %s\n",
				presentation.MetricValue(fmt.Sprintf("%10s", formatSize(summary.SizeBytes))),
				presentation.Primary(summary.Label),
			))
			explainBuilder.WriteString(fmt.Sprintf("%s %s\n", presentation.Muted(fmt.Sprintf("%13s", "Risk:")), presentation.Risk(summary.Risk)))
			explainBuilder.WriteString(fmt.Sprintf("%s %s\n", presentation.Muted(fmt.Sprintf("%13s", "Why:")), presentation.Muted(summary.Why)))
			explainBuilder.WriteString(fmt.Sprintf("%s %s\n", presentation.Muted(fmt.Sprintf("%13s", "Action:")), presentation.Method(summary.Action)))
		}
		return explainBuilder.String()
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

	var b strings.Builder

	b.WriteString(presentation.Brand("◉ HEIMDALL"))
	b.WriteString("\n\n")
	b.WriteString(presentation.Title("Disk scan"))
	b.WriteString("\n")
	b.WriteString(presentation.Divider("──────────────────────────────────────────────────────────────────"))
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("%s%s\n", presentation.Muted(fmt.Sprintf("%-21s", "Path scanned")), presentation.Path(s.RootPath)))
	b.WriteString(fmt.Sprintf("%s%s\n", presentation.Muted(fmt.Sprintf("%-21s", "Total size")), presentation.MetricValue(formatSize(totalSize))))
	b.WriteString(fmt.Sprintf("%s%s\n", presentation.Muted(fmt.Sprintf("%-21s", "Files scanned")), presentation.Primary(formatCount(s.NumFiles))))
	b.WriteString(fmt.Sprintf("%s%s\n", presentation.Muted(fmt.Sprintf("%-21s", "Directories scanned")), presentation.Primary(formatCount(s.NumDirectories))))
	b.WriteString(fmt.Sprintf("%s%s\n\n", presentation.Muted(fmt.Sprintf("%-21s", "Skipped paths")), presentation.Primary(formatCount(len(s.Warnings)))))

	b.WriteString(presentation.Title("Largest directories:"))
	b.WriteString("\n")
	dirsAdded := 0
	for _, dir := range s.GetLargestEntries(dirLimit+1, "dir") {
		if dir == nil || dir.Path == s.RootPath {
			continue
		}
		b.WriteString(fmt.Sprintf("  %s   %s\n",
			presentation.MetricValue(fmt.Sprintf("%10s", formatSize(dir.TotSize))),
			presentation.Path(dir.Path),
		))
		dirsAdded++
		if dirsAdded >= dirLimit {
			break
		}
	}

	b.WriteString("\n")
	b.WriteString(presentation.Title("Largest files:"))
	b.WriteString("\n")
	for _, file := range s.GetLargestEntries(fileLimit, "file") {
		if file == nil {
			continue
		}
		b.WriteString(fmt.Sprintf("  %s   %s\n",
			presentation.MetricValue(fmt.Sprintf("%10s", formatSize(file.TotSize))),
			presentation.Path(file.Path),
		))
	}

	if len(s.Warnings) == 0 {
		return b.String()
	}

	b.WriteString("\n")
	b.WriteString(presentation.Title("Warnings:"))
	b.WriteString("\n")
	for _, warning := range s.Warnings {
		b.WriteString(fmt.Sprintf("  %s %s %s %s\n",
			presentation.Warning("!"),
			presentation.Warning(fmt.Sprintf("%-16s", warning.Type)),
			presentation.Path(warning.Path),
			presentation.Muted("("+warning.Message+")"),
		))
	}

	return b.String()
}
