package blockchain

import (
    "context"
    "fmt"
    "math/big"
    "sync"
    "time"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

// Client represents a connection to an EVM chain
type Client struct {
    chainID   *big.Int
    chainName string
    rpcURL    string
    wsURL     string
    client    *ethclient.Client
    wsClient  *ethclient.Client
    mu        sync.RWMutex
}

// NewClient creates a new blockchain client
func NewClient(chainName string, chainID *big.Int, rpcURL string, wsURL string) (*Client, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to RPC %s: %w", rpcURL, err)
    }
    
    // Verify chain ID
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    networkID, err := client.ChainID(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get chain ID: %w", err)
    }
    
    if networkID.Cmp(chainID) != 0 {
        return nil, fmt.Errorf("chain ID mismatch: expected %s, got %s", chainID, networkID)
    }
    
    c := &Client{
        chainID:   chainID,
        chainName: chainName,
        rpcURL:    rpcURL,
        wsURL:     wsURL,
        client:    client,
    }
    
    // WebSocket connection is optional
    if wsURL != "" {
        wsClient, err := ethclient.Dial(wsURL)
        if err != nil {
            // Log warning but don't fail
            fmt.Printf("Warning: Failed to connect to WebSocket for %s: %v\n", chainName, err)
        } else {
            c.wsClient = wsClient
        }
    }
    
    return c, nil
}

// GetClient returns the HTTP client
func (c *Client) GetClient() *ethclient.Client {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.client
}

// GetWSClient returns the WebSocket client
func (c *Client) GetWSClient() *ethclient.Client {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.wsClient
}

// GetChainID returns the chain ID
func (c *Client) GetChainID() *big.Int {
    return c.chainID
}

// GetChainName returns the chain name
func (c *Client) GetChainName() string {
    return c.chainName
}

// IsWebSocketAvailable checks if WebSocket connection is available
func (c *Client) IsWebSocketAvailable() bool {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.wsClient != nil
}

// GetBlockNumber returns the latest block number
func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
    return c.client.BlockNumber(ctx)
}

// GetBalance returns the ETH balance of an address
func (c *Client) GetBalance(ctx context.Context, address common.Address) (*big.Int, error) {
    return c.client.BalanceAt(ctx, address, nil)
}

// Close closes all client connections
func (c *Client) Close() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    if c.client != nil {
        c.client.Close()
    }
    if c.wsClient != nil {
        c.wsClient.Close()
    }
}