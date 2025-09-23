package styles

import (
	"fmt"
	
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
)

// Color palette
var (
	PrimaryColor   = lipgloss.Color("#00D4AA")
	SecondaryColor = lipgloss.Color("#5A67D8")
	AccentColor    = lipgloss.Color("#F56565")
	SuccessColor   = lipgloss.Color("#48BB78")
	WarningColor   = lipgloss.Color("#ED8936")
	ErrorColor     = lipgloss.Color("#F56565")
	InfoColor      = lipgloss.Color("#4299E1")
	
	// Text colors
	TextPrimary   = lipgloss.Color("#E2E8F0")
	TextSecondary = lipgloss.Color("#A0AEC0")
	TextMuted     = lipgloss.Color("#718096")
	
	// Background colors
	BgPrimary   = lipgloss.Color("#1A202C")
	BgSecondary = lipgloss.Color("#2D3748")
	BgAccent    = lipgloss.Color("#4A5568")
)

// Style definitions
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgPrimary)

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Background(BgPrimary).
			Bold(true).
			Padding(0, 1).
			MarginBottom(1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Background(BgPrimary).
			Italic(true).
			MarginBottom(1)

	// Header styles
	HeaderStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgSecondary).
			Bold(true).
			Padding(0, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(AccentColor)

	// Box styles
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentColor).
			Padding(1, 2).
			Margin(1)

	BoxTitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Align(lipgloss.Center)

	// Status styles
	InfoStyle = lipgloss.NewStyle().
			Foreground(InfoColor).
			Background(BgPrimary)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Background(BgPrimary)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Background(BgPrimary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Background(BgPrimary).
			Bold(true)

	// Input styles
	InputStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgSecondary).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(SecondaryColor)

	InputFocusedStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(BgSecondary).
				Padding(0, 1).
				Border(lipgloss.NormalBorder()).
				BorderForeground(PrimaryColor)

	// Button styles
	ButtonStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(SecondaryColor).
			Padding(0, 2).
			Margin(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(SecondaryColor)

	ButtonActiveStyle = lipgloss.NewStyle().
				Foreground(BgPrimary).
				Background(PrimaryColor).
				Padding(0, 2).
				Margin(0, 1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(PrimaryColor).
				Bold(true)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().
				Foreground(TextPrimary).
				Background(BgSecondary).
				Bold(true).
				Padding(0, 1).
				Border(lipgloss.NormalBorder()).
				BorderForeground(AccentColor)

	TableCellStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Background(BgPrimary).
			Padding(0, 1).
			Border(lipgloss.NormalBorder()).
			BorderForeground(TextMuted)

	TableSelectedStyle = lipgloss.NewStyle().
				Foreground(BgPrimary).
				Background(PrimaryColor).
				Padding(0, 1).
				Border(lipgloss.NormalBorder()).
				BorderForeground(PrimaryColor).
				Bold(true)

	// Chart styles
	ChartBarStyle = lipgloss.NewStyle().
			Background(PrimaryColor)

	ChartAxisStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	ChartLabelStyle = lipgloss.NewStyle().
			Foreground(TextPrimary).
			Bold(true)

	// Footer styles
	FooterStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			Background(BgSecondary).
			Padding(0, 1).
			Align(lipgloss.Center)

	// Help styles
	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(SecondaryColor).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(TextSecondary)

	// Loading styles
	LoadingStyle = lipgloss.NewStyle().
			Foreground(InfoColor).
			Background(BgPrimary).
			Italic(true)

	// Metric styles
	MetricValueStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true).
				Align(lipgloss.Right)

	MetricLabelStyle = lipgloss.NewStyle().
				Foreground(TextSecondary)

	MetricChangePositiveStyle = lipgloss.NewStyle().
					Foreground(SuccessColor).
					Bold(true)

	MetricChangeNegativeStyle = lipgloss.NewStyle().
					Foreground(ErrorColor).
					Bold(true)
)

// Init initializes styles based on configuration
func Init() {
	if viper.GetBool("ui.no_color") {
		// Disable colors for accessibility
		disableColors()
	}
}

// disableColors removes all colors from styles
func disableColors() {
	// Reset all colors to default terminal colors
	PrimaryColor = lipgloss.Color("")
	SecondaryColor = lipgloss.Color("")
	AccentColor = lipgloss.Color("")
	SuccessColor = lipgloss.Color("")
	WarningColor = lipgloss.Color("")
	ErrorColor = lipgloss.Color("")
	InfoColor = lipgloss.Color("")
	
	// Update styles to use no colors
	TitleStyle = TitleStyle.Foreground(lipgloss.Color(""))
	SubtitleStyle = SubtitleStyle.Foreground(lipgloss.Color(""))
	// ... update other styles as needed
}

// GetProgressBar returns a styled progress bar
func GetProgressBar(percent int, width int) string {
	filled := int(float64(width) * float64(percent) / 100.0)
	empty := width - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	
	return lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Render(bar)
}

// FormatCurrency formats a number as currency
func FormatCurrency(value float64) string {
	if value >= 1e9 {
		return MetricValueStyle.Render(fmt.Sprintf("$%.2fB", value/1e9))
	} else if value >= 1e6 {
		return MetricValueStyle.Render(fmt.Sprintf("$%.2fM", value/1e6))
	} else if value >= 1e3 {
		return MetricValueStyle.Render(fmt.Sprintf("$%.2fK", value/1e3))
	}
	return MetricValueStyle.Render(fmt.Sprintf("$%.2f", value))
}

// FormatPercentage formats a number as percentage with color coding
func FormatPercentage(value float64) string {
	if value > 0 {
		return MetricChangePositiveStyle.Render(fmt.Sprintf("+%.2f%%", value))
	} else if value < 0 {
		return MetricChangeNegativeStyle.Render(fmt.Sprintf("%.2f%%", value))
	}
	return MetricValueStyle.Render("0.00%")
}

// GetSparkline generates a sparkline from data points
func GetSparkline(data []float64) string {
	if len(data) == 0 {
		return ""
	}
	
	// Sparkline characters from low to high
	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	
	// Find min and max values
	min, max := data[0], data[0]
	for _, v := range data {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	
	// Handle edge case where all values are the same
	if max == min {
		result := ""
		for range data {
			result += string(chars[len(chars)/2])
		}
		return lipgloss.NewStyle().Foreground(PrimaryColor).Render(result)
	}
	
	// Generate sparkline
	var result string
	for _, v := range data {
		// Normalize value to 0-1 range
		normalized := (v - min) / (max - min)
		// Map to character index
		index := int(normalized * float64(len(chars)-1))
		if index >= len(chars) {
			index = len(chars) - 1
		}
		result += string(chars[index])
	}
	
	return lipgloss.NewStyle().Foreground(PrimaryColor).Render(result)
}

// GetChangeIndicator returns a colored arrow indicating change direction
func GetChangeIndicator(value float64) string {
	if value > 0 {
		return lipgloss.NewStyle().Foreground(SuccessColor).Render("↗")
	} else if value < 0 {
		return lipgloss.NewStyle().Foreground(ErrorColor).Render("↘")
	}
	return lipgloss.NewStyle().Foreground(TextMuted).Render("→")
}

// GetStatusIndicator returns a status indicator dot
func GetStatusIndicator(status string) string {
	switch status {
	case "healthy", "online", "active":
		return lipgloss.NewStyle().Foreground(SuccessColor).Render("●")
	case "warning", "delayed":
		return lipgloss.NewStyle().Foreground(WarningColor).Render("●")
	case "error", "offline", "failed":
		return lipgloss.NewStyle().Foreground(ErrorColor).Render("●")
	default:
		return lipgloss.NewStyle().Foreground(InfoColor).Render("●")
	}
}

// GetProgressRing returns a progress ring (pie chart style)
func GetProgressRing(percent int) string {
	chars := []string{"○", "◔", "◑", "◕", "●"}
	index := int(float64(percent) / 25.0)
	if index >= len(chars) {
		index = len(chars) - 1
	}
	
	if percent >= 75 {
		return lipgloss.NewStyle().Foreground(SuccessColor).Render(chars[index])
	} else if percent >= 50 {
		return lipgloss.NewStyle().Foreground(WarningColor).Render(chars[index])
	} else if percent >= 25 {
		return lipgloss.NewStyle().Foreground(InfoColor).Render(chars[index])
	}
	return lipgloss.NewStyle().Foreground(ErrorColor).Render(chars[index])
}

// CreateASCIIChart creates a simple ASCII bar chart
func CreateASCIIChart(data map[string]float64, width int, height int) string {
	if len(data) == 0 {
		return "No data available"
	}
	
	// Find max value for scaling
	var max float64
	for _, v := range data {
		if v > max {
			max = v
		}
	}
	
	var result []string
	for name, value := range data {
		barLength := int((value / max) * float64(width))
		bar := ""
		
		// Create the bar
		for i := 0; i < barLength; i++ {
			bar += "█"
		}
		for i := barLength; i < width; i++ {
			bar += "░"
		}
		
		// Format the line
		valueStr := FormatCurrency(value)
		line := fmt.Sprintf("%-12s %s %s", 
			name[:min(len(name), 12)], 
			lipgloss.NewStyle().Foreground(PrimaryColor).Render(bar), 
			valueStr)
		result = append(result, line)
		
		if len(result) >= height {
			break
		}
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, result...)
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}