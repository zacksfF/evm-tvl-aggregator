package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/styles"
)

// KeyMap defines key bindings
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Left    key.Binding
	Right   key.Binding
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Refresh, k.Help, k.Quit},
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
}

// DashboardModel represents the main dashboard model
type DashboardModel struct {
	ctx        context.Context
	apiClient  *APIClient
	width      int
	height     int
	ready      bool
	
	// Data
	tvlData       *models.TVLResponse
	protocolsData *models.ProtocolsResponse
	chainsData    *models.ChainsResponse
	statsData     *models.StatsResponse
	
	// UI state
	selectedView int
	loading      bool
	lastUpdate   time.Time
	showHelp     bool
	err          error
	
	// Components
	help help.Model
}

// Init initializes the dashboard
func (d *DashboardModel) Init() tea.Cmd {
	d.help = help.New()
	return tea.Batch(
		d.loadAllData(),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

// Update handles messages
func (d *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			d.loading = true
			cmds = append(cmds, d.loadAllData())
		case key.Matches(msg, keys.Left):
			d.selectedView = max(0, d.selectedView-1)
		case key.Matches(msg, keys.Right):
			d.selectedView = min(2, d.selectedView+1)
		}
		
	case TickMsg:
		if !d.loading {
			d.loading = true
			cmds = append(cmds, d.loadAllData())
		}
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}))
		
	case DataLoadedMsg:
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
		
	case ErrorMsg:
		d.loading = false
		d.err = msg.Err
	}
	
	return d, tea.Batch(cmds...)
}

// View renders the dashboard
func (d *DashboardModel) View() string {
	if !d.ready {
		return "Initializing TVL Dashboard..."
	}
	
	if d.showHelp {
		return d.help.View(keys)
	}
	
	// Header
	title := styles.TitleStyle.Render("📈 TVL Aggregator Dashboard")
	
	// Status indicator
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
	
	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		title,
		lipgloss.NewStyle().Width(d.width-lipgloss.Width(title)-lipgloss.Width(statusText)).Render(""),
		statusText,
	)
	
	// Main content
	var content string
	if d.err != nil {
		content = styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", d.err.Error()))
	} else if d.loading {
		content = styles.LoadingStyle.Render("Loading TVL data...")
	} else {
		content = d.renderDashboard()
	}
	
	// Footer
	helpText := styles.HelpDescStyle.Render("? help • r refresh • q quit")
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		content,
		"",
		helpText,
	)
}

// renderDashboard renders the main dashboard content
func (d *DashboardModel) renderDashboard() string {
	if d.tvlData == nil || d.statsData == nil {
		return styles.InfoStyle.Render("No data available")
	}
	
	// Main metrics
	totalTVL := styles.FormatCurrency(d.tvlData.TotalTVL)
	protocolCount := fmt.Sprintf("%d", d.statsData.ProtocolCount)
	chainCount := fmt.Sprintf("%d", d.statsData.ChainCount)
	
	metrics := lipgloss.JoinHorizontal(
		lipgloss.Top,
		d.renderMetricBox("💰 Total TVL", totalTVL),
		d.renderMetricBox("🔗 Protocols", protocolCount),
		d.renderMetricBox("⛓️ Chains", chainCount),
	)
	
	// Protocol list
	var protocols []string
	if d.protocolsData != nil {
		for _, protocol := range d.protocolsData.Protocols {
			tvl := "N/A"
			if protocolTVL, exists := d.tvlData.Protocols[protocol.Name]; exists {
				tvl = styles.FormatCurrency(protocolTVL.TotalUSD)
			}
			protocols = append(protocols, fmt.Sprintf("%-20s %20s", protocol.Name, tvl))
		}
	}
	
	protocolsContent := "No protocols available"
	if len(protocols) > 0 {
		protocolsContent = lipgloss.JoinVertical(lipgloss.Left, protocols...)
	}
	
	protocolsBox := styles.BoxStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			styles.BoxTitleStyle.Render("📊 Protocol TVL"),
			"",
			protocolsContent,
		),
	)
	
	return lipgloss.JoinVertical(lipgloss.Left, metrics, "", protocolsBox)
}

// renderMetricBox renders a metric box
func (d *DashboardModel) renderMetricBox(title, value string) string {
	return styles.BoxStyle.
		Width(25).
		Height(5).
		Render(
			lipgloss.JoinVertical(lipgloss.Center,
				styles.MetricLabelStyle.Render(title),
				styles.MetricValueStyle.Render(value),
			),
		)
}

// loadAllData loads all dashboard data
func (d *DashboardModel) loadAllData() tea.Cmd {
	return tea.Batch(
		LoadDataCmd(d.apiClient, "tvl"),
		LoadDataCmd(d.apiClient, "protocols"),
		LoadDataCmd(d.apiClient, "chains"),
		LoadDataCmd(d.apiClient, "stats"),
	)
}

// MonitorModel represents a protocol monitoring view
type MonitorModel struct {
	ctx          context.Context
	apiClient    *APIClient
	protocol     string
	chain        string
	width        int
	height       int
	ready        bool
	loading      bool
	protocolData *models.ProtocolTVLResponse
	lastUpdate   time.Time
	err          error
}

// Init initializes the monitor
func (m *MonitorModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadProtocolData(),
		tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}),
	)
}

// Update handles messages for monitor
func (m *MonitorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "r":
			m.loading = true
			cmds = append(cmds, m.loadProtocolData())
		}
		
	case TickMsg:
		if !m.loading {
			m.loading = true
			cmds = append(cmds, m.loadProtocolData())
		}
		cmds = append(cmds, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		}))
		
	case DataLoadedMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		
		if msg.Type == "protocol" {
			if data, ok := msg.Data.(*models.ProtocolTVLResponse); ok {
				m.protocolData = data
			}
		}
		
	case ErrorMsg:
		m.loading = false
		m.err = msg.Err
	}
	
	return m, tea.Batch(cmds...)
}

// View renders the monitor
func (m *MonitorModel) View() string {
	if !m.ready {
		return "Initializing monitor..."
	}
	
	title := styles.TitleStyle.Render(fmt.Sprintf("📊 Monitoring: %s", m.protocol))
	
	var content string
	if m.err != nil {
		content = styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error()))
	} else if m.loading {
		content = styles.LoadingStyle.Render("Loading protocol data...")
	} else if m.protocolData != nil {
		totalTVL := styles.FormatCurrency(m.protocolData.TotalTVL)
		content = fmt.Sprintf("Protocol: %s\nTotal TVL: %s\nLast Update: %s",
			m.protocolData.Protocol,
			totalTVL,
			m.lastUpdate.Format("15:04:05"))
	} else {
		content = "No data available"
	}
	
	help := styles.HelpDescStyle.Render("r refresh • q quit")
	
	return lipgloss.JoinVertical(lipgloss.Left, title, "", content, "", help)
}

// loadProtocolData loads protocol-specific data
func (m *MonitorModel) loadProtocolData() tea.Cmd {
	return func() tea.Msg {
		data, err := m.apiClient.GetProtocolTVL(m.protocol)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return DataLoadedMsg{
			Data: data,
			Type: "protocol",
		}
	}
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