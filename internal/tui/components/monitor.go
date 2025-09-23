package components

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/styles"
)

// Monitor represents a protocol monitoring view
type Monitor struct {
	ctx        context.Context
	apiClient  *tui.APIClient
	protocol   string
	chain      string
	width      int
	height     int
	ready      bool
	
	// Data
	protocolData *models.ProtocolTVLResponse
	loading      bool
	lastUpdate   time.Time
	err          error
}

// NewMonitor creates a new monitor model
func NewMonitor(apiClient *tui.APIClient, ctx context.Context, protocol, chain string) (*Monitor, error) {
	return &Monitor{
		ctx:       ctx,
		apiClient: apiClient,
		protocol:  protocol,
		chain:     chain,
		loading:   true,
	}, nil
}

// Init initializes the monitor
func (m *Monitor) Init() tea.Cmd {
	return tea.Batch(
		m.loadProtocolData(),
		tickCmd(),
	)
}

// Update handles messages
func (m *Monitor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		
	case tui.TickMsg:
		if !m.loading {
			m.loading = true
			cmds = append(cmds, m.loadProtocolData())
		}
		cmds = append(cmds, tickCmd())
		
	case tui.DataLoadedMsg:
		m.loading = false
		m.lastUpdate = time.Now()
		
		if msg.Type == "protocol" {
			if data, ok := msg.Data.(*models.ProtocolTVLResponse); ok {
				m.protocolData = data
			}
		}
		
	case tui.ErrorMsg:
		m.loading = false
		m.err = msg.Err
	}
	
	return m, tea.Batch(cmds...)
}

// View renders the monitor
func (m *Monitor) View() string {
	if !m.ready {
		return "Initializing monitor..."
	}
	
	title := styles.TitleStyle.Render(fmt.Sprintf("📊 Monitoring: %s", m.protocol))
	
	var content string
	if m.protocolData == nil {
		content = styles.LoadingStyle.Render("Loading protocol data...")
	} else {
		content = m.renderProtocolDetails()
	}
	
	status := m.renderStatus()
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		content,
		status,
	)
}

// renderProtocolDetails renders the protocol details
func (m *Monitor) renderProtocolDetails() string {
	data := m.protocolData
	
	// Main metrics
	totalTVL := styles.FormatCurrency(data.TotalTVL)
	timestamp := data.Timestamp.Format("15:04:05")
	
	metrics := tui.RenderBox(
		lipgloss.JoinVertical(lipgloss.Left,
			styles.MetricLabelStyle.Render("Protocol: "+data.Protocol),
			styles.MetricValueStyle.Render("TVL: "+totalTVL),
			styles.MetricLabelStyle.Render("Updated: "+timestamp),
		),
		"📈 Protocol Metrics",
	)
	
	// Chain breakdown
	var chainRows []string
	for chain, chainData := range data.Chains {
		if m.chain == "" || m.chain == chain {
			chainTVL := styles.FormatCurrency(chainData.TotalUSD)
			assetCount := fmt.Sprintf("%d assets", len(chainData.Assets))
			
			row := fmt.Sprintf("%-20s %20s %15s", chain, chainTVL, assetCount)
			chainRows = append(chainRows, styles.TableCellStyle.Render(row))
		}
	}
	
	chainsHeader := styles.TableHeaderStyle.Render(fmt.Sprintf("%-20s %20s %15s", "Chain", "TVL", "Assets"))
	chainsContent := lipgloss.JoinVertical(lipgloss.Left, chainsHeader)
	if len(chainRows) > 0 {
		chainsContent = lipgloss.JoinVertical(lipgloss.Left, chainsHeader, lipgloss.JoinVertical(lipgloss.Left, chainRows...))
	}
	
	chainsBox := tui.RenderBox(chainsContent, "⛓️ Chain Breakdown")
	
	return lipgloss.JoinVertical(lipgloss.Left, metrics, chainsBox)
}

// renderStatus renders the status bar
func (m *Monitor) renderStatus() string {
	var status string
	if m.err != nil {
		status = styles.ErrorStyle.Render(fmt.Sprintf("Error: %s", m.err.Error()))
	} else if m.loading {
		status = styles.LoadingStyle.Render("Loading...")
	} else {
		status = styles.InfoStyle.Render(fmt.Sprintf("Last updated: %s", m.lastUpdate.Format("15:04:05")))
	}
	
	help := styles.HelpDescStyle.Render("r refresh • q quit")
	
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		status,
		lipgloss.NewStyle().Width(m.width-lipgloss.Width(status)-lipgloss.Width(help)).Render(""),
		help,
	)
}

// loadProtocolData loads protocol-specific data
func (m *Monitor) loadProtocolData() tea.Cmd {
	return func() tea.Msg {
		data, err := m.apiClient.GetProtocolTVL(m.protocol)
		if err != nil {
			return tui.ErrorMsg{Err: err}
		}
		return tui.DataLoadedMsg{
			Data: data,
			Type: "protocol",
		}
	}
}