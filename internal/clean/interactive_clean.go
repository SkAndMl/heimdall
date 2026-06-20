package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SkAndMl/heimdall/internal/scan"
	"github.com/SkAndMl/heimdall/internal/trash"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	findings        []scan.Finding
	cursor          int
	selected        map[int]bool
	done            bool
	totSizeSelected int64
	trashErr        string
}

func initialModel(findings []scan.Finding) model {
	return model{
		findings: findings,
		selected: make(map[int]bool),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.done {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			default:
				return m, nil
			}
		}

		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.findings)-1 {
				m.cursor++
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y":
			m.done = true
			paths := make([]string, 0)
			for idx, ok := range m.selected {
				if ok {
					paths = append(paths, m.findings[idx].Path)
				}
			}
			if err := trash.MoveToTrash(paths...); err != nil {
				m.trashErr = err.Error()
			}
		case "N":
			return m, tea.Quit
		case " ":
			size := m.findings[m.cursor].Size
			m.selected[m.cursor] = !m.selected[m.cursor]
			if m.selected[m.cursor] {
				m.totSizeSelected += size
			} else {
				m.totSizeSelected -= size
			}
		}
	}

	return m, nil
}

func (m model) View() string {
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

	if m.done {
		var selectedCount int
		var selectedSize int64
		for i, finding := range m.findings {
			if !m.selected[i] {
				continue
			}
			selectedCount++
			selectedSize += finding.Size
		}

		var b strings.Builder
		if m.trashErr != "" {
			b.WriteString("Cleanup failed.\n\n")
			b.WriteString(fmt.Sprintf("Error: %s\n\n", m.trashErr))
		} else {
			b.WriteString("Cleanup complete.\n\n")
			b.WriteString(fmt.Sprintf("Moved %d item(s) to Trash.\n", selectedCount))
			b.WriteString(fmt.Sprintf("Freed: %s\n\n", formatSize(selectedSize)))
		}
		b.WriteString("Press q to exit.\n")
		return b.String()
	}

	categoryForPath := func(path string) string {
		clean := filepath.Clean(path)
		base := filepath.Base(clean)
		lowerPath := strings.ToLower(clean)
		lowerBase := strings.ToLower(base)

		switch {
		case lowerBase == "__pycache__":
			return "python_cache"
		case lowerBase == "node_modules":
			return "node_modules"
		case lowerBase == ".venv" || lowerBase == "venv" || strings.HasSuffix(lowerPath, string(filepath.Separator)+"pyvenv.cfg"):
			return "python_virtual_environment"
		case strings.HasSuffix(lowerBase, ".dmg"),
			strings.HasSuffix(lowerBase, ".pkg"),
			strings.HasSuffix(lowerBase, ".mpkg"),
			strings.HasSuffix(lowerBase, ".msi"),
			strings.HasSuffix(lowerBase, ".exe"),
			strings.HasSuffix(lowerBase, ".deb"),
			strings.HasSuffix(lowerBase, ".rpm"),
			strings.HasSuffix(lowerBase, ".appimage"):
			return "installer"
		case strings.HasSuffix(lowerBase, ".tar.gz"),
			strings.HasSuffix(lowerBase, ".tar.bz2"),
			strings.HasSuffix(lowerBase, ".tgz"),
			strings.HasSuffix(lowerBase, ".tar.xz"),
			strings.HasSuffix(lowerBase, ".zip"),
			strings.HasSuffix(lowerBase, ".7z"),
			strings.HasSuffix(lowerBase, ".rar"),
			strings.HasSuffix(lowerBase, ".tar"),
			strings.HasSuffix(lowerBase, ".gz"),
			strings.HasSuffix(lowerBase, ".bz2"),
			strings.HasSuffix(lowerBase, ".xz"):
			return "archive"
		default:
			return ""
		}
	}

	labelForPath := func(path string) string {
		clean := filepath.Clean(path)
		base := filepath.Base(clean)
		parent := filepath.Base(filepath.Dir(clean))

		switch base {
		case "node_modules", ".venv", "venv":
			if parent != "." && parent != string(filepath.Separator) {
				return filepath.Join(parent, base)
			}
		}
		return base
	}

	riskForPath := func(path string) string {
		category := categoryForPath(path)
		if explanation, ok := scan.CategoryExplanations[category]; ok {
			return explanation.Risk
		}
		return "Review recommended"
	}

	const (
		nameWidth = 30
		sizeWidth = 8
	)

	var selectedSize int64
	action := "move to Trash"
	for i, finding := range m.findings {
		if !m.selected[i] {
			continue
		}
		selectedSize += finding.Size
		if riskForPath(finding.Path) == "Review recommended" {
			action = "select manually"
		}
	}

	var b strings.Builder
	b.WriteString("Select cleanup candidates\n\n")

	const maxVisibleRows = 18
	start := 0
	if m.cursor >= maxVisibleRows {
		start = m.cursor - maxVisibleRows + 1
	}
	end := start + maxVisibleRows
	if end > len(m.findings) {
		end = len(m.findings)
	}

	if start > 0 {
		b.WriteString(fmt.Sprintf("  ... %d more above\n", start))
	}

	for i := start; i < end; i++ {
		finding := m.findings[i]
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}

		label := labelForPath(finding.Path)
		if len(label) > nameWidth {
			label = label[:nameWidth-3] + "..."
		}

		b.WriteString(fmt.Sprintf("%s%s %-*s %-*s %s\n",
			cursor,
			checkbox,
			nameWidth,
			label,
			sizeWidth,
			formatSize(finding.Size),
			riskForPath(finding.Path),
		))
	}

	if end < len(m.findings) {
		b.WriteString(fmt.Sprintf("  ... %d more below\n", len(m.findings)-end))
	}

	b.WriteString(fmt.Sprintf("\nSelected: %s\n", formatSize(selectedSize)))
	b.WriteString(fmt.Sprintf("Action: %s\n", action))
	b.WriteString("Continue? y/N\n")
	b.WriteString("\nup/down or j/k: move | space: toggle | y: confirm | q: quit\n")

	return b.String()
}

func RunInteractiveClean(findings []scan.Finding) error {
	p := tea.NewProgram(initialModel(findings), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	return nil
}
