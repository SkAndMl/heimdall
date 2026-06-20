package clean

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SkAndMl/heimdall/internal/scan"
)

type Options struct {
	Path        string
	DryRun      bool
	Interactive bool
}

func Clean(args Options) (string, error) {
	scanner, err := scan.NewScanner(args.Path, -1)
	if err != nil {
		return "", fmt.Errorf("scan cleanup candidates: %w", err)
	}

	if args.DryRun {
		return DryRunReport(scanner.Categories), nil
	}

	if args.Interactive {
		findings := make([]scan.Finding, 0)
		for _, categoryFindings := range scanner.Categories {
			findings = append(findings, categoryFindings...)
		}
		sort.Slice(findings, func(i, j int) bool {
			return findings[i].Size > findings[j].Size
		})
		if err := RunInteractiveClean(findings); err != nil {
			return "", err
		}
	}

	return "", nil
}

func DryRunReport(categories map[string][]scan.Finding) string {
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

	pluralize := func(count int, singular string, plural string) string {
		if count == 1 {
			return fmt.Sprintf("%d %s", count, singular)
		}
		return fmt.Sprintf("%d %s", count, plural)
	}

	wrapText := func(text string, width int) []string {
		words := strings.Fields(text)
		if len(words) == 0 {
			return []string{""}
		}

		lines := make([]string, 0)
		current := words[0]
		for _, word := range words[1:] {
			if len(current)+1+len(word) > width {
				lines = append(lines, current)
				current = word
				continue
			}
			current += " " + word
		}
		lines = append(lines, current)
		return lines
	}

	type cleanupSummary struct {
		section string
		label   string
		detail  string
		action  string
		size    int64
		order   int
	}

	summaries := make([]cleanupSummary, 0)

	if findings := categories["python_cache"]; len(findings) > 0 {
		var size int64
		for _, finding := range findings {
			size += finding.Size
		}
		summaries = append(summaries, cleanupSummary{
			section: "Safe candidates",
			label:   "Python bytecode cache",
			detail:  pluralize(len(findings), "__pycache__ directory", "__pycache__ directories"),
			action:  "move to Trash",
			size:    size,
			order:   0,
		})
	}

	archiveFindings := append([]scan.Finding{}, categories["installer"]...)
	archiveFindings = append(archiveFindings, categories["archive"]...)
	if len(archiveFindings) > 0 {
		var size int64
		extensionCounts := make(map[string]int)
		for _, finding := range archiveFindings {
			size += finding.Size
			extension := strings.TrimPrefix(strings.ToUpper(filepath.Ext(finding.Path)), ".")
			if extension == "GZ" && strings.HasSuffix(strings.ToLower(finding.Path), ".tar.gz") {
				extension = "TAR.GZ"
			}
			if extension == "" {
				extension = "file"
			}
			extensionCounts[extension]++
		}

		extensions := make([]string, 0, len(extensionCounts))
		for extension := range extensionCounts {
			extensions = append(extensions, extension)
		}
		sort.Strings(extensions)

		detailParts := make([]string, 0, len(extensions))
		for _, extension := range extensions {
			detailParts = append(detailParts, pluralize(extensionCounts[extension], extension+" file", extension+" files"))
		}

		summaries = append(summaries, cleanupSummary{
			section: "Usually safe",
			label:   "Installers and archives",
			detail:  strings.Join(detailParts, ", "),
			action:  "move to Trash",
			size:    size,
			order:   1,
		})
	}

	if findings := categories["python_virtual_environment"]; len(findings) > 0 {
		var size int64
		for _, finding := range findings {
			size += finding.Size
		}
		summaries = append(summaries, cleanupSummary{
			section: "Review recommended",
			label:   "Python virtual environments",
			detail:  pluralize(len(findings), "environment detected", "environments detected"),
			action:  "select manually",
			size:    size,
			order:   2,
		})
	}

	if findings := categories["node_modules"]; len(findings) > 0 {
		var size int64
		for _, finding := range findings {
			size += finding.Size
		}
		summaries = append(summaries, cleanupSummary{
			section: "Review recommended",
			label:   "Node dependencies",
			detail:  pluralize(len(findings), "node_modules directory", "node_modules directories"),
			action:  "select manually",
			size:    size,
			order:   3,
		})
	}

	if findings := categories["huggingface_cache"]; len(findings) > 0 {
		var size int64
		for _, finding := range findings {
			size += finding.Size
		}
		summaries = append(summaries, cleanupSummary{
			section: "Review recommended",
			label:   "Hugging Face cache",
			detail:  pluralize(len(findings), "cache path detected", "cache paths detected"),
			action:  "select manually",
			size:    size,
			order:   4,
		})
	}

	sort.SliceStable(summaries, func(i, j int) bool {
		return summaries[i].order < summaries[j].order
	})

	var b strings.Builder
	b.WriteString("Heimdall Cleanup Plan\n\n")

	totalSelectable := int64(0)
	for _, summary := range summaries {
		totalSelectable += summary.size
	}

	const (
		sizeWidth   = 8
		detailPad   = "            "
		detailWidth = 68
	)

	for _, section := range []string{"Safe candidates", "Usually safe", "Review recommended"} {
		wroteSection := false
		for _, summary := range summaries {
			if summary.section != section {
				continue
			}
			if !wroteSection {
				b.WriteString(section + ":\n")
				wroteSection = true
			}
			b.WriteString(fmt.Sprintf("  %-*s %s\n", sizeWidth, formatSize(summary.size), summary.label))
			for _, line := range wrapText(summary.detail, detailWidth) {
				b.WriteString(fmt.Sprintf("%s%s\n", detailPad, line))
			}
			b.WriteString(fmt.Sprintf("%sAction: %s\n\n", detailPad, summary.action))
		}
	}

	b.WriteString(fmt.Sprintf("Total safely selectable: %s\n\n", formatSize(totalSelectable)))
	b.WriteString("No files were changed. Run:\n")
	b.WriteString("  heimdall clean ~ --interactive")

	return b.String()
}
