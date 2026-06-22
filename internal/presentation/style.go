package presentation

import (
	"os"

	"github.com/SkAndMl/heimdall/internal/categories"
	"github.com/charmbracelet/lipgloss"
)

var (
	brandStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F2B950"))
	titleStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#E8EEF8"))
	dividerStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#26334B"))
	mutedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#91A0B6"))
	primaryStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#E8EEF8"))
	goldStyle       = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F2B950"))
	cyanStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#69D9E7"))
	mintStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#67D7A7"))
	amberStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A524"))
	coralStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#F27474"))
	focusStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#69D9E7"))
	unselectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#91A0B6"))
)

func colorEnabled() bool {
	return os.Getenv("NO_COLOR") == ""
}

func render(style lipgloss.Style, text string) string {
	if !colorEnabled() {
		return text
	}
	return style.Render(text)
}

func Brand(text string) string {
	return render(brandStyle, text)
}

func Title(text string) string {
	return render(titleStyle, text)
}

func Divider(text string) string {
	return render(dividerStyle, text)
}

func Muted(text string) string {
	return render(mutedStyle, text)
}

func Primary(text string) string {
	return render(primaryStyle, text)
}

func MetricValue(text string) string {
	return render(goldStyle, text)
}

func Path(text string) string {
	return render(cyanStyle, text)
}

func Method(text string) string {
	return render(cyanStyle, text)
}

func Success(text string) string {
	return render(mintStyle, text)
}

func Warning(text string) string {
	return render(amberStyle, text)
}

func Danger(text string) string {
	return render(coralStyle, text)
}

func Focus(text string) string {
	return render(focusStyle, text)
}

func Selected(text string) string {
	return render(goldStyle, text)
}

func Unselected(text string) string {
	return render(unselectedStyle, text)
}

func Risk(text string) string {
	switch text {
	case categories.RiskUsuallySafe:
		return Success(text)
	case categories.RiskReviewRecommended:
		return Warning(text)
	default:
		return Primary(text)
	}
}
