package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/viper"
	"github.com/zacksfF/evm-tvl-aggregator/internal/tui/styles"
)

// AppMode represents the application mode
type AppMode int

const (
	DashboardMode AppMode = iota
	MonitorMode
	ProtocolMode
)

// App represents the main TUI application
type App struct {
	mode      AppMode
	protocol  string
	chain     string
	apiClient *APIClient
	program   *tea.Program
	model     tea.Model
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewApp creates a new TUI application
func NewApp() (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	apiURL := viper.GetString("api.url")
	apiClient, err := NewAPIClient(apiURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create API client: %w", err)
	}

	app := &App{
		mode:      DashboardMode,
		apiClient: apiClient,
		ctx:       ctx,
		cancel:    cancel,
	}

	return app, nil
}

// SetMode sets the application mode
func (a *App) SetMode(mode AppMode) {
	a.mode = mode
}

// SetProtocol sets the protocol to monitor
func (a *App) SetProtocol(protocol string) {
	a.protocol = protocol
}

// SetChain sets the chain filter
func (a *App) SetChain(chain string) {
	a.chain = chain
}

// Run starts the TUI application
func (a *App) Run() error {
	defer a.cancel()

	// Create the appropriate model based on mode
	var initialModel tea.Model
	
	switch a.mode {
	case DashboardMode:
		initialModel = &DashboardModel{
			ctx:       a.ctx,
			apiClient: a.apiClient,
			loading:   true,
		}
	case MonitorMode:
		initialModel = &MonitorModel{
			ctx:       a.ctx,
			apiClient: a.apiClient,
			protocol:  a.protocol,
			chain:     a.chain,
			loading:   true,
		}
	default:
		initialModel = &DashboardModel{
			ctx:       a.ctx,
			apiClient: a.apiClient,
			loading:   true,
		}
	}

	// Configure the program
	a.program = tea.NewProgram(
		initialModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Setup styles
	styles.Init()

	// Run the program
	_, err := a.program.Run()
	return err
}

// Quit gracefully shuts down the application
func (a *App) Quit() {
	if a.program != nil {
		a.program.Quit()
	}
	a.cancel()
}


// TickCmd returns a command that sends a tick message
func TickCmd() tea.Cmd {
	refreshInterval := viper.GetDuration("ui.refresh_interval")
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Second
	}

	return tea.Tick(refreshInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// LoadDataCmd returns a command to load data
func LoadDataCmd(apiClient *APIClient, dataType string) tea.Cmd {
	return func() tea.Msg {
		var data interface{}
		var err error

		switch dataType {
		case "tvl":
			data, err = apiClient.GetTotalTVL()
		case "protocols":
			data, err = apiClient.GetProtocols()
		case "chains":
			data, err = apiClient.GetChains()
		case "stats":
			data, err = apiClient.GetStats()
		default:
			return ErrorMsg{Err: fmt.Errorf("unknown data type: %s", dataType)}
		}

		if err != nil {
			return ErrorMsg{Err: err}
		}

		return DataLoadedMsg{
			Data: data,
			Type: dataType,
		}
	}
}
