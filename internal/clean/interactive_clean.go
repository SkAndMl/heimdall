package clean

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/SkAndMl/heimdall/internal/categories"
	"github.com/SkAndMl/heimdall/internal/presentation"
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

	actionLabel := func(action string) string {
		if action == categories.ActionSelectManually {
			return "review selected items before moving to Trash"
		}
		return "move selected items to Trash"
	}

	const (
		nameWidth = 50
		sizeWidth = 10
	)

	var reclaimableSize int64
	var selectedSize int64
	action := categories.ActionMoveToTrash
	for i, finding := range m.findings {
		reclaimableSize += finding.Size
		if !m.selected[i] {
			continue
		}
		selectedSize += finding.Size
		if riskForFinding(finding) == categories.RiskReviewRecommended {
			action = categories.ActionSelectManually
		}
	}

	var b strings.Builder
	b.WriteString(presentation.Brand("◉ HEIMDALL"))
	b.WriteString("\n\n")
	b.WriteString(presentation.Title("Cleanup candidates"))
	b.WriteString("\n")
	b.WriteString(presentation.Divider("──────────────────────────────────────────────────────────────────"))
	b.WriteString("\n\n")

	if len(m.findings) == 0 {
		b.WriteString(presentation.Primary("No cleanup candidates found yet."))
		b.WriteString("\n\n")
		b.WriteString(presentation.Muted("Run a scan to inspect developer artifacts, installers, and caches."))
		b.WriteString("\n")
		b.WriteString("\n")
		b.WriteString(presentation.Muted("q Quit"))
		b.WriteString("\n")
		return b.String()
	}

	b.WriteString(fmt.Sprintf("%s %s\n",
		presentation.Muted(fmt.Sprintf("%-26s", "Potentially reclaimable")),
		presentation.MetricValue(fmt.Sprintf("%16s", formatSize(reclaimableSize))),
	))
	b.WriteString(fmt.Sprintf("%s %s\n\n",
		presentation.Muted(fmt.Sprintf("%-26s", "Selected")),
		presentation.MetricValue(fmt.Sprintf("%16s", formatSize(selectedSize))),
	))

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
		b.WriteString(presentation.Muted(fmt.Sprintf("  · %d more above", start)))
		b.WriteString("\n")
	}

	for i := start; i < end; i++ {
		finding := m.findings[i]
		checkbox := presentation.Unselected("○")
		if m.selected[i] {
			checkbox = presentation.Selected("✓")
		}

		cursor := "  "
		if i == m.cursor {
			cursor = presentation.Focus("›") + " "
		}

		label := labelForPath(finding.Path)
		if len(label) > nameWidth {
			label = "..." + label[len(label)-(nameWidth-3):]
		}

		b.WriteString(fmt.Sprintf("%s%s  %s %s   %s\n",
			cursor,
			checkbox,
			presentation.Path(fmt.Sprintf("%-*s", nameWidth, label)),
			presentation.MetricValue(fmt.Sprintf("%*s", sizeWidth, formatSize(finding.Size))),
			presentation.Risk(riskForFinding(finding)),
		))
	}

	if end < len(m.findings) {
		b.WriteString(presentation.Muted(fmt.Sprintf("  · %d more below", len(m.findings)-end)))
		b.WriteString("\n")
	}

	b.WriteString(fmt.Sprintf("\n%s %s\n", presentation.Muted("Method:"), presentation.Method(actionLabel(action))))
	b.WriteString("Continue? y/N\n")
	b.WriteString("\n")
	b.WriteString(presentation.Muted("↑↓ or j/k Move   Space Select   y Confirm   n/q Quit"))
	b.WriteString("\n")
	if m.noneSelectedYesPressed {
		b.WriteString("\n")
		b.WriteString(presentation.Warning("Select at least one candidate before continuing."))
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
