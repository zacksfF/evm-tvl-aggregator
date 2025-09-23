package components

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/zacksfF/evm-tvl-aggregator/internal/tui"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/styles"
)

// KeyMap defines key bindings for the dashboard
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Refresh   key.Binding
	Help      key.Binding
	Quit      key.Binding
	Select    key.Binding
	Back      key.Binding
	Search    key.Binding
	Sort      key.Binding
	Filter    key.Binding
	Export    key.Binding
	Pause     key.Binding
	Fullscreen key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Select, k.Back, k.Refresh, k.Pause},
		{k.Search, k.Sort, k.Filter, k.Export},
		{k.Fullscreen, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Back: key.NewBinding(
		key.WithKeys("backspace"),
		key.WithHelp("backspace", "back"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Filter: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter"),
	),
	Export: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "export"),
	),
	Pause: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "pause/resume"),
	),
	Fullscreen: key.NewBinding(
		key.WithKeys("F11"),
		key.WithHelp("F11", "fullscreen"),
	),
}

// Dashboard represents the main dashboard model
type Dashboard struct {
	ctx       context.Context
	apiClient *tui.APIClient
	width     int
	height    int
	ready     bool

	// Data
	tvlData       *models.TVLResponse
	protocolsData *models.ProtocolsResponse
	chainsData    *models.ChainsResponse
	statsData     *models.StatsResponse

	// UI state
	selectedView int // 0: overview, 1: protocols, 2: chains
	selectedItem int
	loading      bool
	lastUpdate   time.Time
	showHelp     bool
	paused       bool
	searchMode   bool
	searchQuery  string

	// Components
	help help.Model

	// Error handling
	err error
}

// NewDashboard creates a new dashboard model
func NewDashboard(apiClient *tui.APIClient, ctx context.Context) (*Dashboard, error) {
	d := &Dashboard{
		ctx:       ctx,
		apiClient: apiClient,
		help:      help.New(),
		ready:     false,
		loading:   true,
	}

	return d, nil
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return tea.Batch(
		d.loadAllData(),
		tickCmd(),
	)
}

// Update handles messages
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.ready = true

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return d, tea.Quit

		case key.Matches(msg, keys.Help):
			d.showHelp = !d.showHelp

		case key.Matches(msg, keys.Refresh):
			if !d.paused {
				d.loading = true
				cmds = append(cmds, d.loadAllData())
			}

		case key.Matches(msg, keys.Pause):
			d.paused = !d.paused

		case key.Matches(msg, keys.Search):
			d.searchMode = !d.searchMode
			d.searchQuery = ""

		case key.Matches(msg, keys.Sort):
			// Toggle sort order (implementation would depend on current view)
			
		case key.Matches(msg, keys.Filter):
			// Show filter options

		case key.Matches(msg, keys.Export):
			// Export current data to CSV

		case key.Matches(msg, keys.Up):
			if d.selectedView == 1 && len(d.protocolsData.Protocols) > 0 {
				d.selectedItem = max(0, d.selectedItem-1)
			} else if d.selectedView == 2 && len(d.chainsData.Chains) > 0 {
				d.selectedItem = max(0, d.selectedItem-1)
			}

		case key.Matches(msg, keys.Down):
			if d.selectedView == 1 && len(d.protocolsData.Protocols) > 0 {
				d.selectedItem = min(len(d.protocolsData.Protocols)-1, d.selectedItem+1)
			} else if d.selectedView == 2 && len(d.chainsData.Chains) > 0 {
				d.selectedItem = min(len(d.chainsData.Chains)-1, d.selectedItem+1)
			}

		case key.Matches(msg, keys.Left):
			d.selectedView = max(0, d.selectedView-1)
			d.selectedItem = 0

		case key.Matches(msg, keys.Right):
			d.selectedView = min(2, d.selectedView+1)
			d.selectedItem = 0
		}

	case tui.TickMsg:
		if !d.loading && !d.paused {
			d.loading = true
			cmds = append(cmds, d.loadAllData())
		}
		cmds = append(cmds, tui.TickCmd())

	case tui.DataLoadedMsg:
		d.loading = false
		d.lastUpdate = time.Now()

		switch msg.Type {
		case "tvl":
			if data, ok := msg.Data.(*models.TVLResponse); ok {
				d.tvlData = data
			}
		case "protocols":
			if data, ok := msg.Data.(*models.ProtocolsResponse); ok {
				d.protocolsData = data
			}
		case "chains":
			if data, ok := msg.Data.(*models.ChainsResponse); ok {
				d.chainsData = data
			}
		case "stats":
			if data, ok := msg.Data.(*models.StatsResponse); ok {
				d.statsData = data
			}
		}

	case tui.ErrorMsg:
		d.loading = false
		d.err = msg.Err
	}

	return d, tea.Batch(cmds...)
}

// View renders the dashboard
func (d *Dashboard) View() string {
	if !d.ready {
		return "Initializing..."
	}

	if d.showHelp {
		return d.help.View(keys)
	}

	// Header
	header := d.renderHeader()

	// Main content based on selected view
	var content string
	switch d.selectedView {
	case 0:
		content = d.renderOverview()
	case 1:
		content = d.renderProtocols()
	case 2:
		content = d.renderChains()
	}

	// Footer
	footer := d.renderFooter()

	// Calculate available height for content
	headerHeight := lipgloss.Height(header)
	footerHeight := lipgloss.Height(footer)
	contentHeight := d.height - headerHeight - footerHeight - 2

	// Combine all parts
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		lipgloss.NewStyle().Height(contentHeight).Render(content),
		footer,
	)
}

// renderHeader renders the dashboard header
func (d *Dashboard) renderHeader() string {
	title := styles.TitleStyle.Render("📈 TVL Aggregator Dashboard")

	status := "●"
	statusColor := styles.SuccessColor
	if d.loading {
		status = "◐"
		statusColor = styles.WarningColor
	} else if d.err != nil {
		status = "●"
		statusColor = styles.ErrorColor
	}

	statusText := lipgloss.NewStyle().
		Foreground(statusColor).
		Render(status + " " + d.lastUpdate.Format("15:04:05"))

	tabs := []string{"Overview", "Protocols", "Chains"}
	var tabsRendered []string

	for i, tab := range tabs {
		if i == d.selectedView {
			tabsRendered = append(tabsRendered, styles.ButtonActiveStyle.Render(tab))
		} else {
			tabsRendered = append(tabsRendered, styles.ButtonStyle.Render(tab))
		}
	}

	tabsLine := lipgloss.JoinHorizontal(lipgloss.Left, tabsRendered...)

	headerContent := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			lipgloss.NewStyle().Width(d.width-lipgloss.Width(title)-lipgloss.Width(statusText)).Render(""),
			statusText,
		),
		tabsLine,
	)

	return styles.HeaderStyle.Width(d.width).Render(headerContent)
}

// renderOverview renders the overview tab
func (d *Dashboard) renderOverview() string {
	if d.tvlData == nil || d.statsData == nil {
		return styles.LoadingStyle.Render("Loading overview data...")
	}

	// Main metrics with sparklines and indicators
	totalTVL := styles.FormatCurrency(d.tvlData.TotalTVL)
	protocolCount := fmt.Sprintf("%d", d.statsData.ProtocolCount)
	chainCount := fmt.Sprintf("%d", d.statsData.ChainCount)

	// Generate mock historical data for sparklines (in real app, this would come from API)
	tvlHistory := []float64{d.tvlData.TotalTVL * 0.95, d.tvlData.TotalTVL * 0.98, d.tvlData.TotalTVL * 1.02, d.tvlData.TotalTVL}
	tvlSparkline := styles.GetSparkline(tvlHistory)
	tvlChange := 2.5 // Mock 24h change %
	tvlIndicator := styles.GetChangeIndicator(tvlChange)
	
	// Connection status
	connStatus := styles.GetStatusIndicator("healthy")

	metrics := lipgloss.JoinHorizontal(
		lipgloss.Top,
		tui.RenderBox(
			lipgloss.JoinVertical(lipgloss.Center,
				styles.MetricLabelStyle.Render("Total TVL"),
				totalTVL,
				lipgloss.JoinHorizontal(lipgloss.Center, 
					tvlIndicator, " ", 
					styles.FormatPercentage(tvlChange), " ",
					tvlSparkline,
				),
			),
			"💰 Total Value Locked",
		),
		tui.RenderBox(
			lipgloss.JoinVertical(lipgloss.Center,
				styles.MetricLabelStyle.Render("Protocols"),
				styles.MetricValueStyle.Render(protocolCount),
				styles.GetProgressRing(75) + " Active",
			),
			"🔗 Active Protocols",
		),
		tui.RenderBox(
			lipgloss.JoinVertical(lipgloss.Center,
				styles.MetricLabelStyle.Render("Chains"),
				styles.MetricValueStyle.Render(chainCount),
				connStatus + " " + lipgloss.NewStyle().Foreground(styles.TextSecondary).Render("Connected"),
			),
			"⛓️ Supported Chains",
		),
	)

	// Enhanced protocol distribution chart
	if len(d.tvlData.Protocols) > 0 {
		// Convert protocol data to map for chart
		protocolTVLs := make(map[string]float64)
		for name, data := range d.tvlData.Protocols {
			protocolTVLs[name] = data.TotalUSD
		}
		
		// Create ASCII bar chart
		chart := styles.CreateASCIIChart(protocolTVLs, 50, 8)
		chartBox := tui.RenderBox(chart, "📊 Protocol TVL Distribution")

		// Add top movers section
		topMovers := d.renderTopMovers()

		return lipgloss.JoinVertical(lipgloss.Left, 
			metrics, 
			lipgloss.JoinHorizontal(lipgloss.Top, chartBox, topMovers),
		)
	}

	return metrics
}

// renderTopMovers renders the top movers section
func (d *Dashboard) renderTopMovers() string {
	if d.tvlData == nil {
		return ""
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		styles.BoxTitleStyle.Render("🚀 Top Movers (24h)"),
		"",
		// Mock data - in real app, this would come from API
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.GetChangeIndicator(5.2), " ",
			lipgloss.NewStyle().Foreground(styles.TextPrimary).Render("Uniswap V2"), " ",
			styles.FormatPercentage(5.2),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.GetChangeIndicator(-2.1), " ",
			lipgloss.NewStyle().Foreground(styles.TextPrimary).Render("Aave V3"), " ",
			styles.FormatPercentage(-2.1),
		),
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.GetChangeIndicator(1.8), " ",
			lipgloss.NewStyle().Foreground(styles.TextPrimary).Render("Compound"), " ",
			styles.FormatPercentage(1.8),
		),
	)

	return tui.RenderBox(content, "")
}

// renderProtocols renders the protocols tab
func (d *Dashboard) renderProtocols() string {
	if d.protocolsData == nil {
		return styles.LoadingStyle.Render("Loading protocols data...")
	}

	if len(d.protocolsData.Protocols) == 0 {
		return styles.InfoStyle.Render("No protocols found")
	}

	var rows []string
	for i, protocol := range d.protocolsData.Protocols {
		style := styles.TableCellStyle
		if i == d.selectedItem {
			style = styles.TableSelectedStyle
		}

		chains := lipgloss.JoinHorizontal(lipgloss.Left, protocol.Chains...)
		row := style.Render(fmt.Sprintf("%-20s %-10s %-30s %s",
			protocol.Name,
			protocol.Type,
			protocol.Description,
			chains,
		))
		rows = append(rows, row)
	}

	header := styles.TableHeaderStyle.Render(fmt.Sprintf("%-20s %-10s %-30s %s",
		"Name", "Type", "Description", "Chains"))

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderChains renders the chains tab
func (d *Dashboard) renderChains() string {
	if d.chainsData == nil {
		return styles.LoadingStyle.Render("Loading chains data...")
	}

	if len(d.chainsData.Chains) == 0 {
		return styles.InfoStyle.Render("No chains found")
	}

	var rows []string
	for i, chain := range d.chainsData.Chains {
		style := styles.TableCellStyle
		if i == d.selectedItem {
			style = styles.TableSelectedStyle
		}

		tvl := "N/A"
		if d.tvlData != nil && d.tvlData.Chains != nil {
			if chainTVL, exists := d.tvlData.Chains[chain]; exists {
				tvl = styles.FormatCurrency(chainTVL)
			}
		}

		row := style.Render(fmt.Sprintf("%-20s %s", chain, tvl))
		rows = append(rows, row)
	}

	header := styles.TableHeaderStyle.Render(fmt.Sprintf("%-20s %s", "Chain", "TVL"))

	return lipgloss.JoinVertical(lipgloss.Left, header, lipgloss.JoinVertical(lipgloss.Left, rows...))
}

// renderFooter renders the dashboard footer
func (d *Dashboard) renderFooter() string {
	// Left side - status and info
	var leftStatus []string
	
	if d.err != nil {
		leftStatus = append(leftStatus, styles.ErrorStyle.Render(fmt.Sprintf("🔴 Error: %s", d.err.Error())))
	} else if d.loading {
		leftStatus = append(leftStatus, styles.LoadingStyle.Render("🔄 Loading..."))
	} else {
		leftStatus = append(leftStatus, styles.InfoStyle.Render(fmt.Sprintf("✓ Updated: %s", d.lastUpdate.Format("15:04:05"))))
	}
	
	// Add pause indicator
	if d.paused {
		leftStatus = append(leftStatus, styles.WarningStyle.Render("⏸ Paused"))
	}
	
	// Add search indicator
	if d.searchMode {
		leftStatus = append(leftStatus, styles.InfoStyle.Render(fmt.Sprintf("🔍 Search: %s", d.searchQuery)))
	}
	
	statusText := lipgloss.JoinHorizontal(lipgloss.Left, leftStatus...)
	
	// Right side - keyboard shortcuts
	shortcuts := []string{
		"? help", 
		"space pause", 
		"/ search", 
		"r refresh", 
		"q quit",
	}
	help := styles.HelpDescStyle.Render(lipgloss.JoinHorizontal(lipgloss.Left, shortcuts...))

	// Center spacing
	availableWidth := d.width - lipgloss.Width(statusText) - lipgloss.Width(help) - 4
	if availableWidth < 0 {
		availableWidth = 0
	}
	
	footer := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusText,
		lipgloss.NewStyle().Width(availableWidth).Render(""),
		help,
	)

	return styles.FooterStyle.Width(d.width).Render(footer)
}

// loadAllData loads all dashboard data
func (d *Dashboard) loadAllData() tea.Cmd {
	return tea.Batch(
		tui.LoadDataCmd(d.apiClient, "tvl"),
		tui.LoadDataCmd(d.apiClient, "protocols"),
		tui.LoadDataCmd(d.apiClient, "chains"),
		tui.LoadDataCmd(d.apiClient, "stats"),
	)
}

// tickCmd returns a command that sends a tick message
func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tui.TickMsg(t)
	})
}

// Helper functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
