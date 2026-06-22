package clean

import (
	"errors"
	"strings"
	"testing"

	"github.com/SkAndMl/heimdall/internal/categories"
	"github.com/SkAndMl/heimdall/internal/scan"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInteractiveModelConfirmsSelectedFindings(t *testing.T) {
	findings := []scan.Finding{
		{Path: "/tmp/archive.zip", Size: 1024, Category: categories.Archive},
		{Path: "/tmp/app.dmg", Size: 2048, Category: categories.Installer},
	}

	m := initialModel("/tmp", findings)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	m = updated.(model)

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("y")})
	m = updated.(model)

	selection := m.selection()
	if !selection.Confirmed {
		t.Fatal("selection.Confirmed = false, want true")
	}
	if len(selection.Findings) != 1 {
		t.Fatalf("selected findings = %d, want 1", len(selection.Findings))
	}
	if selection.Findings[0].Path != findings[0].Path {
		t.Fatalf("selected path = %q, want %q", selection.Findings[0].Path, findings[0].Path)
	}
}

func TestFinishInteractiveCleanMovesConfirmedSelection(t *testing.T) {
	selection := interactiveSelection{
		Confirmed: true,
		Findings: []scan.Finding{
			{Path: "/tmp/archive.zip", Size: 1024, Category: categories.Archive},
			{Path: "/tmp/app.dmg", Size: 2048, Category: categories.Installer},
		},
	}

	var movedPaths []string
	report, err := finishInteractiveClean(selection, func(paths ...string) error {
		movedPaths = append(movedPaths, paths...)
		return nil
	})

	if err != nil {
		t.Fatalf("finishInteractiveClean() error = %v, want nil", err)
	}
	if len(movedPaths) != 2 {
		t.Fatalf("moved paths = %d, want 2", len(movedPaths))
	}
	if movedPaths[0] != "/tmp/archive.zip" || movedPaths[1] != "/tmp/app.dmg" {
		t.Fatalf("moved paths = %#v, want selected paths in order", movedPaths)
	}
	if !strings.Contains(report, "Moved 2 artifacts to Trash") {
		t.Fatalf("report = %q, want moved count", report)
	}
	if !strings.Contains(report, "3.0 KB") {
		t.Fatalf("report = %q, want selected size", report)
	}
	if strings.Contains(report, "Freed") {
		t.Fatalf("report = %q, did not want overpromising freed-space copy", report)
	}
}

func TestInteractiveViewShowsDesignedSummary(t *testing.T) {
	findings := []scan.Finding{
		{Path: "/tmp/archive.zip", Size: 1024, Category: categories.Archive},
		{Path: "/tmp/app.dmg", Size: 2048, Category: categories.Installer},
	}

	m := initialModel("/tmp", findings)
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace, Runes: []rune(" ")})
	m = updated.(model)

	view := m.View()
	if !strings.Contains(view, "◉ HEIMDALL") {
		t.Fatalf("View() = %q, want product header", view)
	}
	if !strings.Contains(view, "Potentially reclaimable") {
		t.Fatalf("View() = %q, want reclaimable summary", view)
	}
	if !strings.Contains(view, "Selected") || !strings.Contains(view, "1.0 KB") {
		t.Fatalf("View() = %q, want selected-size summary", view)
	}
	if !strings.Contains(view, "✓") {
		t.Fatalf("View() = %q, want selected marker", view)
	}
	if !strings.Contains(view, "Space Select") {
		t.Fatalf("View() = %q, want keyboard hint", view)
	}
}

func TestInteractiveViewShowsEmptyState(t *testing.T) {
	m := initialModel("/tmp", nil)

	view := m.View()
	if !strings.Contains(view, "No cleanup candidates found yet.") {
		t.Fatalf("View() = %q, want empty state", view)
	}
	if !strings.Contains(view, "q Quit") {
		t.Fatalf("View() = %q, want quit hint", view)
	}
}

func TestFinishInteractiveCleanSkipsUnconfirmedSelection(t *testing.T) {
	called := false
	report, err := finishInteractiveClean(interactiveSelection{}, func(paths ...string) error {
		called = true
		return nil
	})

	if err != nil {
		t.Fatalf("finishInteractiveClean() error = %v, want nil", err)
	}
	if report != "" {
		t.Fatalf("report = %q, want empty", report)
	}
	if called {
		t.Fatal("moveToTrash was called for unconfirmed selection")
	}
}

func TestFinishInteractiveCleanReturnsTrashError(t *testing.T) {
	trashErr := errors.New("trash unavailable")
	selection := interactiveSelection{
		Confirmed: true,
		Findings:  []scan.Finding{{Path: "/tmp/archive.zip", Size: 1024, Category: categories.Archive}},
	}

	report, err := finishInteractiveClean(selection, func(paths ...string) error {
		return trashErr
	})

	if err == nil {
		t.Fatal("finishInteractiveClean() error = nil, want trash error")
	}
	if !errors.Is(err, trashErr) {
		t.Fatalf("finishInteractiveClean() error = %v, want wrapped trash error", err)
	}
	if report != "" {
		t.Fatalf("report = %q, want empty", report)
	}
}
