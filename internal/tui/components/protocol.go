package components

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui"
)

// ProtocolView represents a detailed protocol view
type ProtocolView struct {
	ctx       context.Context
	apiClient *tui.APIClient
	protocol  string
	width     int
	height    int
	ready     bool
}

// NewProtocolView creates a new protocol view model
func NewProtocolView(apiClient *tui.APIClient, ctx context.Context, protocol string) (*ProtocolView, error) {
	return &ProtocolView{
		ctx:       ctx,
		apiClient: apiClient,
		protocol:  protocol,
	}, nil
}

// Init initializes the protocol view
func (p *ProtocolView) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (p *ProtocolView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height
		p.ready = true
		
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return p, tea.Quit
		}
	}
	
	return p, nil
}

// View renders the protocol view
func (p *ProtocolView) View() string {
	if !p.ready {
		return "Initializing protocol view..."
	}
	
	return "Protocol view for: " + p.protocol
}