package blockchain

import (
    "context"
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

// EventFilter helps build event queries
type EventFilter struct {
    FromBlock *big.Int
    ToBlock   *big.Int
    Addresses []common.Address
    Topics    [][]common.Hash
}

// EventSignature generates event signature hash
func EventSignature(eventName string) common.Hash {
    return crypto.Keccak256Hash([]byte(eventName))
}

// Common ERC20 Events
var (
    TransferEventSig = EventSignature("Transfer(address,address,uint256)")
    ApprovalEventSig = EventSignature("Approval(address,address,uint256)")
)

// Common DeFi Events
var (
    DepositEventSig  = EventSignature("Deposit(address,uint256)")
    WithdrawEventSig = EventSignature("Withdraw(address,uint256)")
    SwapEventSig     = EventSignature("Swap(address,address,uint256,uint256)")
)

// GetLogs fetches logs from the blockchain
func (c *Client) GetLogs(ctx context.Context, filter EventFilter) ([]types.Log, error) {
    query := ethereum.FilterQuery{
        FromBlock: filter.FromBlock,
        ToBlock:   filter.ToBlock,
        Addresses: filter.Addresses,
        Topics:    filter.Topics,
    }
    
    logs, err := c.client.FilterLogs(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to filter logs: %w", err)
    }
    
    return logs, nil
}

// SubscribeLogs subscribes to new logs (requires WebSocket)
func (c *Client) SubscribeLogs(ctx context.Context, filter EventFilter) (<-chan types.Log, ethereum.Subscription, error) {
    if !c.IsWebSocketAvailable() {
        return nil, nil, fmt.Errorf("WebSocket connection not available for chain %s", c.chainName)
    }
    
    query := ethereum.FilterQuery{
        Addresses: filter.Addresses,
        Topics:    filter.Topics,
    }
    
    logs := make(chan types.Log)
    sub, err := c.wsClient.SubscribeFilterLogs(ctx, query, logs)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to subscribe to logs: %w", err)
    }
    
    return logs, sub, nil
}

// ParseTransferEvent parses ERC20 Transfer event
type TransferEvent struct {
    From   common.Address
    To     common.Address
    Amount *big.Int
}

func ParseTransferEvent(log types.Log) (*TransferEvent, error) {
    if len(log.Topics) != 3 {
        return nil, fmt.Errorf("invalid transfer event topics")
    }
    
    event := &TransferEvent{
        From:   common.HexToAddress(log.Topics[1].Hex()),
        To:     common.HexToAddress(log.Topics[2].Hex()),
        Amount: new(big.Int).SetBytes(log.Data),
    }
    
    return event, nil
}

// EventListener listens for events across multiple contracts
type EventListener struct {
    client   *Client
    filters  []EventFilter
    handlers map[common.Hash]func(types.Log) error
}

// NewEventListener creates a new event listener
func NewEventListener(client *Client) *EventListener {
    return &EventListener{
        client:   client,
        filters:  []EventFilter{},
        handlers: make(map[common.Hash]func(types.Log) error),
    }
}

// AddEventHandler adds a handler for a specific event signature
func (el *EventListener) AddEventHandler(eventSig common.Hash, handler func(types.Log) error) {
    el.handlers[eventSig] = handler
}

// AddFilter adds a filter for events to listen to
func (el *EventListener) AddFilter(filter EventFilter) {
    el.filters = append(el.filters, filter)
}

// Start starts listening for events
func (el *EventListener) Start(ctx context.Context) error {
    if !el.client.IsWebSocketAvailable() {
        return fmt.Errorf("WebSocket not available for real-time event listening")
    }
    
    for _, filter := range el.filters {
        logs, sub, err := el.client.SubscribeLogs(ctx, filter)
        if err != nil {
            return err
        }
        
        go el.handleLogs(ctx, logs, sub)
    }
    
    return nil
}

func (el *EventListener) handleLogs(ctx context.Context, logs <-chan types.Log, sub ethereum.Subscription) {
    for {
        select {
        case err := <-sub.Err():
            fmt.Printf("Subscription error: %v\n", err)
            return
            
        case log := <-logs:
            if len(log.Topics) > 0 {
                if handler, exists := el.handlers[log.Topics[0]]; exists {
                    if err := handler(log); err != nil {
                        fmt.Printf("Error handling log: %v\n", err)
                    }
                }
            }
            
        case <-ctx.Done():
            sub.Unsubscribe()
            return
        }
    }
}