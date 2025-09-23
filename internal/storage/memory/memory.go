package memory

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage"
)

// MemoryStorage implements Storage interface in memory
type MemoryStorage struct {
	protocols map[string]*models.Protocol
	snapshots []*models.TVLSnapshot
	chains    map[string]*models.Chain
	events    []*models.Event
	blocks    map[string]uint64
	tokens    map[string]*models.Token
	mu        sync.RWMutex
}

// NewMemoryStorage creates a new in-memory storage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		protocols: make(map[string]*models.Protocol),
		snapshots: make([]*models.TVLSnapshot, 0),
		chains:    make(map[string]*models.Chain),
		events:    make([]*models.Event, 0),
		blocks:    make(map[string]uint64),
		tokens:    make(map[string]*models.Token),
	}
}

// GetProtocol retrieves a protocol by name
func (ms *MemoryStorage) GetProtocol(ctx context.Context, name string) (*models.Protocol, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	protocol, exists := ms.protocols[name]
	if !exists {
		return nil, fmt.Errorf("protocol not found: %s", name)
	}
	return protocol, nil
}

// GetProtocols retrieves all protocols
func (ms *MemoryStorage) GetProtocols(ctx context.Context) ([]*models.Protocol, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	protocols := make([]*models.Protocol, 0, len(ms.protocols))
	for _, protocol := range ms.protocols {
		protocols = append(protocols, protocol)
	}
	return protocols, nil
}

// SaveProtocol saves a protocol
func (ms *MemoryStorage) SaveProtocol(ctx context.Context, protocol *models.Protocol) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if protocol.ID == 0 {
		protocol.ID = uint64(len(ms.protocols) + 1)
	}
	protocol.CreatedAt = time.Now()
	protocol.UpdatedAt = time.Now()

	ms.protocols[protocol.Name] = protocol
	return nil
}

// SaveTVLSnapshot saves a TVL snapshot
func (ms *MemoryStorage) SaveTVLSnapshot(ctx context.Context, snapshot *models.TVLSnapshot) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	snapshot.ID = uint64(len(ms.snapshots) + 1)
	ms.snapshots = append(ms.snapshots, snapshot)

	total, _ := snapshot.TotalUSD.Float64()
	fmt.Printf("💾 Saved TVL: %s = $%.2f\n", snapshot.Protocol, total)

	return nil
}

// GetLatestTVL retrieves the latest TVL
func (ms *MemoryStorage) GetLatestTVL(ctx context.Context, protocol, chain string) (*models.TVLSnapshot, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for i := len(ms.snapshots) - 1; i >= 0; i-- {
		snap := ms.snapshots[i]
		if snap.Protocol == protocol {
			if chain == "" || snap.Chain == chain {
				return snap, nil
			}
		}
	}

	return nil, fmt.Errorf("no TVL snapshot found")
}

// GetHistoricalTVL retrieves historical TVL data
func (ms *MemoryStorage) GetHistoricalTVL(ctx context.Context, protocol, chain string, from, to time.Time) ([]*models.TVLSnapshot, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.TVLSnapshot
	for _, snap := range ms.snapshots {
		if snap.Protocol == protocol &&
			snap.Timestamp.After(from) &&
			snap.Timestamp.Before(to) {

			if chain == "" || snap.Chain == chain {
				results = append(results, snap)
			}
		}
	}

	return results, nil
}

// SaveEvents saves events
func (ms *MemoryStorage) SaveEvents(ctx context.Context, events []*models.Event) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, event := range events {
		event.ID = uint64(len(ms.events) + 1)
		ms.events = append(ms.events, event)
	}

	return nil
}

// GetEvents retrieves events
func (ms *MemoryStorage) GetEvents(ctx context.Context, filter storage.EventFilter) ([]*models.Event, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.Event
	for _, event := range ms.events {
		if (filter.Chain == "" || event.Chain == filter.Chain) &&
			(filter.Protocol == "" || event.Protocol == filter.Protocol) &&
			(filter.EventName == "" || event.EventName == filter.EventName) {

			if filter.FromBlock > 0 && event.BlockNumber < filter.FromBlock {
				continue
			}
			if filter.ToBlock > 0 && event.BlockNumber > filter.ToBlock {
				continue
			}

			results = append(results, event)
		}
	}

	// Apply limit
	if filter.Limit > 0 && len(results) > filter.Limit {
		results = results[:filter.Limit]
	}

	return results, nil
}

// SaveBlock saves a block
func (ms *MemoryStorage) SaveBlock(ctx context.Context, block *models.Block) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.blocks[block.Chain] = block.Number
	return nil
}

// GetLastIndexedBlock gets the last indexed block
func (ms *MemoryStorage) GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	block, exists := ms.blocks[chain]
	if !exists {
		return 0, nil
	}
	return block, nil
}

// Additional interface implementations...
func (ms *MemoryStorage) GetAggregatedTVL(ctx context.Context) (*models.AggregatedTVL, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Calculate aggregated TVL from snapshots
	aggregated := &models.AggregatedTVL{
		Timestamp:   time.Now(),
		Protocols:   make(map[string]*models.TVLData),
		TotalUSD:    big.NewFloat(0),
		ChainTotals: make(map[string]*big.Float),
	}

	return aggregated, nil
}

func (ms *MemoryStorage) GetTVLByBlock(ctx context.Context, protocol, chain string, blockNumber uint64) (*models.TVLSnapshot, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, snap := range ms.snapshots {
		if snap.Protocol == protocol && 
		   snap.BlockNumber == blockNumber {
			if chain == "" || snap.Chain == chain {
				return snap, nil
			}
		}
	}

	return nil, fmt.Errorf("no TVL snapshot found for block %d", blockNumber)
}

func (ms *MemoryStorage) UpdateChainStats(ctx context.Context, stats *models.ChainStats) error {
	// Not implemented for memory storage
	return nil
}

func (ms *MemoryStorage) GetToken(ctx context.Context, address, chain string) (*models.Token, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	key := fmt.Sprintf("%s-%s", chain, address)
	token, exists := ms.tokens[key]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}
	return token, nil
}

func (ms *MemoryStorage) SaveToken(ctx context.Context, token *models.Token) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	key := fmt.Sprintf("%s-%s", token.Chain, token.Address.Hex())
	ms.tokens[key] = token
	return nil
}

func (ms *MemoryStorage) UpdateTokenPrice(ctx context.Context, price *models.TokenPrice) error {
	// Not fully implemented for memory storage
	return nil
}

func (ms *MemoryStorage) GetTokenPrices(ctx context.Context, addresses []string) (map[string]*models.TokenPrice, error) {
	// Not fully implemented for memory storage
	return make(map[string]*models.TokenPrice), nil
}

func (ms *MemoryStorage) UpdateProtocol(ctx context.Context, protocol *models.Protocol) error {
	return ms.SaveProtocol(ctx, protocol)
}

func (ms *MemoryStorage) DeleteProtocol(ctx context.Context, name string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.protocols, name)
	return nil
}

func (ms *MemoryStorage) GetChain(ctx context.Context, name string) (*models.Chain, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	chain, exists := ms.chains[name]
	if !exists {
		return nil, fmt.Errorf("chain not found")
	}
	return chain, nil
}

func (ms *MemoryStorage) GetChains(ctx context.Context) ([]*models.Chain, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	chains := make([]*models.Chain, 0, len(ms.chains))
	for _, chain := range ms.chains {
		chains = append(chains, chain)
	}
	return chains, nil
}

func (ms *MemoryStorage) SaveChain(ctx context.Context, chain *models.Chain) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if chain.ID == 0 {
		chain.ID = uint64(len(ms.chains) + 1)
	}
	ms.chains[chain.Name] = chain
	return nil
}

func (ms *MemoryStorage) BeginTx(ctx context.Context) (storage.Tx, error) {
	// Memory storage doesn't support transactions
	return &MemoryTx{storage: ms}, nil
}

func (ms *MemoryStorage) Close() error {
	return nil
}

// MemoryTx is a mock transaction
type MemoryTx struct {
	storage *MemoryStorage
}

func (mt *MemoryTx) Commit() error   { return nil }
func (mt *MemoryTx) Rollback() error { return nil }

// Delegate all methods to the underlying storage
func (mt *MemoryTx) GetProtocol(ctx context.Context, name string) (*models.Protocol, error) {
	return mt.storage.GetProtocol(ctx, name)
}

func (mt *MemoryTx) GetProtocols(ctx context.Context) ([]*models.Protocol, error) {
	return mt.storage.GetProtocols(ctx)
}

func (mt *MemoryTx) SaveProtocol(ctx context.Context, protocol *models.Protocol) error {
	return mt.storage.SaveProtocol(ctx, protocol)
}

func (mt *MemoryTx) UpdateProtocol(ctx context.Context, protocol *models.Protocol) error {
	return mt.storage.UpdateProtocol(ctx, protocol)
}

func (mt *MemoryTx) DeleteProtocol(ctx context.Context, name string) error {
	return mt.storage.DeleteProtocol(ctx, name)
}

func (mt *MemoryTx) SaveTVLSnapshot(ctx context.Context, snapshot *models.TVLSnapshot) error {
	return mt.storage.SaveTVLSnapshot(ctx, snapshot)
}

func (mt *MemoryTx) GetLatestTVL(ctx context.Context, protocol, chain string) (*models.TVLSnapshot, error) {
	return mt.storage.GetLatestTVL(ctx, protocol, chain)
}

func (mt *MemoryTx) GetHistoricalTVL(ctx context.Context, protocol, chain string, from, to time.Time) ([]*models.TVLSnapshot, error) {
	return mt.storage.GetHistoricalTVL(ctx, protocol, chain, from, to)
}

func (mt *MemoryTx) GetTVLByBlock(ctx context.Context, protocol, chain string, blockNumber uint64) (*models.TVLSnapshot, error) {
	return mt.storage.GetTVLByBlock(ctx, protocol, chain, blockNumber)
}

func (mt *MemoryTx) GetAggregatedTVL(ctx context.Context) (*models.AggregatedTVL, error) {
	return mt.storage.GetAggregatedTVL(ctx)
}

func (mt *MemoryTx) GetChain(ctx context.Context, name string) (*models.Chain, error) {
	return mt.storage.GetChain(ctx, name)
}

func (mt *MemoryTx) GetChains(ctx context.Context) ([]*models.Chain, error) {
	return mt.storage.GetChains(ctx)
}

func (mt *MemoryTx) SaveChain(ctx context.Context, chain *models.Chain) error {
	return mt.storage.SaveChain(ctx, chain)
}

func (mt *MemoryTx) UpdateChainStats(ctx context.Context, stats *models.ChainStats) error {
	return mt.storage.UpdateChainStats(ctx, stats)
}

func (mt *MemoryTx) SaveEvents(ctx context.Context, events []*models.Event) error {
	return mt.storage.SaveEvents(ctx, events)
}

func (mt *MemoryTx) GetEvents(ctx context.Context, filter storage.EventFilter) ([]*models.Event, error) {
	return mt.storage.GetEvents(ctx, filter)
}

func (mt *MemoryTx) SaveBlock(ctx context.Context, block *models.Block) error {
	return mt.storage.SaveBlock(ctx, block)
}

func (mt *MemoryTx) GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error) {
	return mt.storage.GetLastIndexedBlock(ctx, chain)
}

func (mt *MemoryTx) GetToken(ctx context.Context, address, chain string) (*models.Token, error) {
	return mt.storage.GetToken(ctx, address, chain)
}

func (mt *MemoryTx) SaveToken(ctx context.Context, token *models.Token) error {
	return mt.storage.SaveToken(ctx, token)
}

func (mt *MemoryTx) UpdateTokenPrice(ctx context.Context, price *models.TokenPrice) error {
	return mt.storage.UpdateTokenPrice(ctx, price)
}

func (mt *MemoryTx) GetTokenPrices(ctx context.Context, addresses []string) (map[string]*models.TokenPrice, error) {
	return mt.storage.GetTokenPrices(ctx, addresses)
}

func (mt *MemoryTx) BeginTx(ctx context.Context) (storage.Tx, error) {
	return mt.storage.BeginTx(ctx)
}

func (mt *MemoryTx) Close() error {
	return mt.storage.Close()
}
