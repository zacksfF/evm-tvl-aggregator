package blockchain

import (
    "context"
    "fmt"
    "math/big"
    "sync"
)

// ChainConfig represents configuration for a blockchain
type ChainConfig struct {
    Name      string
    ChainID   *big.Int
    RPCURL    string
    WSURL     string
    Explorer  string
    NativeToken string
}

// Manager manages multiple blockchain clients
type Manager struct {
    clients map[string]*Client
    configs map[string]*ChainConfig
    mu      sync.RWMutex
}

// NewManager creates a new blockchain manager
func NewManager() *Manager {
    return &Manager{
        clients: make(map[string]*Client),
        configs: make(map[string]*ChainConfig),
    }
}

// AddChain adds a new chain to the manager
func (m *Manager) AddChain(config ChainConfig) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if _, exists := m.clients[config.Name]; exists {
        return fmt.Errorf("chain %s already exists", config.Name)
    }
    
    client, err := NewClient(config.Name, config.ChainID, config.RPCURL, config.WSURL)
    if err != nil {
        return fmt.Errorf("failed to create client for %s: %w", config.Name, err)
    }
    
    m.clients[config.Name] = client
    m.configs[config.Name] = &config
    
    fmt.Printf("Successfully added chain: %s (ID: %s)\n", config.Name, config.ChainID)
    return nil
}

// GetClient returns a client for a specific chain
func (m *Manager) GetClient(chainName string) (*Client, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    client, exists := m.clients[chainName]
    if !exists {
        return nil, fmt.Errorf("client for chain %s not found", chainName)
    }
    return client, nil
}

// GetAllClients returns all clients
func (m *Manager) GetAllClients() map[string]*Client {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    clients := make(map[string]*Client)
    for k, v := range m.clients {
        clients[k] = v
    }
    return clients
}

// GetChainConfig returns configuration for a specific chain
func (m *Manager) GetChainConfig(chainName string) (*ChainConfig, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    config, exists := m.configs[chainName]
    if !exists {
        return nil, fmt.Errorf("config for chain %s not found", chainName)
    }
    return config, nil
}

// GetSupportedChains returns list of supported chain names
func (m *Manager) GetSupportedChains() []string {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    chains := make([]string, 0, len(m.clients))
    for name := range m.clients {
        chains = append(chains, name)
    }
    return chains
}

// HealthCheck checks the health of all connections
func (m *Manager) HealthCheck(ctx context.Context) map[string]bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    health := make(map[string]bool)
    
    for name, client := range m.clients {
        _, err := client.GetBlockNumber(ctx)
        health[name] = err == nil
    }
    
    return health
}

// Close closes all client connections
func (m *Manager) Close() {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    for name, client := range m.clients {
        client.Close()
        fmt.Printf("Closed connection to %s\n", name)
    }
    
    m.clients = make(map[string]*Client)
}