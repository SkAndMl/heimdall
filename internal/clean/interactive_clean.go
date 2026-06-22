package clean

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SkAndMl/heimdall/internal/categories"
	"github.com/SkAndMl/heimdall/internal/scan"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	rootPath               string
	findings               []scan.Finding
	cursor                 int
	selected               map[int]bool
	confirmed              bool
	noneSelectedYesPressed bool
}

type interactiveSelection struct {
	Confirmed bool
	Findings  []scan.Finding
}

func initialModel(rootPath string, findings []scan.Finding) model {
	return model{
		rootPath: rootPath,
		findings: findings,
		selected: make(map[int]bool),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) selectedFindings() []scan.Finding {
	selected := make([]scan.Finding, 0)
	for i, finding := range m.findings {
		if m.selected[i] {
			selected = append(selected, finding)
		}
	}
	return selected
}

func (m model) selection() interactiveSelection {
	return interactiveSelection{
		Confirmed: m.confirmed,
		Findings:  m.selectedFindings(),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	m.noneSelectedYesPressed = false

	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			if len(m.selectedFindings()) == 0 {
				m.noneSelectedYesPressed = true
				return m, nil
			}

			m.confirmed = true
			return m, tea.Quit
		case "N", "n":
			return m, tea.Quit
		case " ":
			if len(m.findings) == 0 {
				return m, nil
			}
			m.selected[m.cursor] = !m.selected[m.cursor]
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

	labelForPath := func(path string) string {
		clean := filepath.Clean(path)
		relPath, err := filepath.Rel(m.rootPath, clean)
		pathToReturn := clean
		if err == nil {
			pathToReturn = relPath
		}
		return pathToReturn
	}

	riskForFinding := func(finding scan.Finding) string {
		if info, ok := categories.Lookup(finding.Category); ok {
			return info.Risk
		}
		return categories.RiskReviewRecommended
	}

	const (
		nameWidth = 54
		sizeWidth = 8
	)

	var selectedSize int64
	action := categories.ActionMoveToTrash
	for i, finding := range m.findings {
		if !m.selected[i] {
			continue
		}
		selectedSize += finding.Size
		if riskForFinding(finding) == categories.RiskReviewRecommended {
			action = categories.ActionSelectManually
		}
	}

	var b strings.Builder
	b.WriteString("Select cleanup candidates\n\n")

	if len(m.findings) == 0 {
		b.WriteString("No cleanup candidates found.\n")
		b.WriteString("\nq: quit\n")
		return b.String()
	}

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
			left := (nameWidth - 3) / 2
			right := nameWidth - 3 - left
			label = label[:left] + "..." + label[len(label)-right:]
		}

		b.WriteString(fmt.Sprintf("%s%s %-*s %-*s %s\n",
			cursor,
			checkbox,
			nameWidth,
			label,
			sizeWidth,
			formatSize(finding.Size),
			riskForFinding(finding),
		))
	}

	if end < len(m.findings) {
		b.WriteString(fmt.Sprintf("  ... %d more below\n", len(m.findings)-end))
	}

	b.WriteString(fmt.Sprintf("\nSelected: %s\n", formatSize(selectedSize)))
	b.WriteString(fmt.Sprintf("Action: %s\n", action))
	b.WriteString("Continue? y/N\n")
	b.WriteString("\nup/down or j/k: move | space: toggle | y: confirm | q: quit\n")
	if m.noneSelectedYesPressed {
		b.WriteString("\nSelect at least one item before continuing.")
	}
	return b.String()
}

func runInteractiveClean(rootPath string, findings []scan.Finding) (interactiveSelection, error) {
	p := tea.NewProgram(initialModel(rootPath, findings), tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return interactiveSelection{}, fmt.Errorf("run interactive clean: %w", err)
	}

	m, ok := finalModel.(model)
	if !ok {
		return interactiveSelection{}, fmt.Errorf("run interactive clean: unexpected model %T", finalModel)
	}
	return m.selection(), nil
}
