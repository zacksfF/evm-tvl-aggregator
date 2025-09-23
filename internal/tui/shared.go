package tui

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/styles"
)

// Common message types for all components
type (
	// DataLoadedMsg indicates that data has been loaded
	DataLoadedMsg struct {
		Data interface{}
		Type string
	}

	// ErrorMsg represents an error
	ErrorMsg struct {
		Err error
	}

	// StatusMsg represents a status update
	StatusMsg struct {
		Status string
		Level  string // info, warn, error
	}

	// TickMsg represents a tick message for periodic updates
	TickMsg time.Time
)

// Helper functions for styling
func RenderTitle(title string) string {
	return styles.TitleStyle.Render(title)
}

func RenderSubtitle(subtitle string) string {
	return styles.SubtitleStyle.Render(subtitle)
}

func RenderStatus(status, level string) string {
	switch level {
	case "error":
		return styles.ErrorStyle.Render(status)
	case "warn":
		return styles.WarningStyle.Render(status)
	default:
		return styles.InfoStyle.Render(status)
	}
}

func RenderBox(content string, title string) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.AccentColor).
		Padding(1, 2).
		Width(50).
		Render(lipgloss.JoinVertical(lipgloss.Left,
			styles.BoxTitleStyle.Render(title),
			"",
			content,
		))
}
