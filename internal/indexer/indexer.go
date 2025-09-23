// internal/indexer/indexer.go
package indexer

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
)

type Indexer struct {
    manager      *blockchain.Manager
    storage      Storage
    processors   map[string]EventProcessor
    config       Config
    wg           sync.WaitGroup
    mu           sync.RWMutex
}

type Config struct {
    BatchSize          uint64
    WorkerCount        int
    BlockConfirmations uint64
    RetryAttempts      int
    RetryDelay         time.Duration
    StartFromBlock     uint64  // 0 means latest - 1000
}

type Storage interface {
    SaveBlock(ctx context.Context, block *Block) error
    GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error)
    SaveEvents(ctx context.Context, events []*Event) error
    GetEvents(ctx context.Context, chain string, from, to uint64) ([]*Event, error)
}

type EventProcessor interface {
    Process(log types.Log) (*Event, error)
    GetEventSignatures() []common.Hash
    GetProtocolName() string
}

func NewIndexer(manager *blockchain.Manager, storage Storage, config Config) *Indexer {
    return &Indexer{
        manager:    manager,
        storage:    storage,
        processors: make(map[string]EventProcessor),
        config:     config,
    }
}

func (idx *Indexer) RegisterProcessor(processor EventProcessor) {
    idx.mu.Lock()
    defer idx.mu.Unlock()
    
    name := processor.GetProtocolName()
    idx.processors[name] = processor
    fmt.Printf("Registered processor for protocol: %s\n", name)
}

func (idx *Indexer) Start(ctx context.Context) error {
    chains := idx.manager.GetSupportedChains()
    
    for _, chain := range chains {
        idx.wg.Add(1)
        go func(chainName string) {
            defer idx.wg.Done()
            
            if err := idx.indexChain(ctx, chainName); err != nil {
                fmt.Printf("Error indexing chain %s: %v\n", chainName, err)
            }
        }(chain)
    }
    
    idx.wg.Wait()
    return nil
}

func (idx *Indexer) indexChain(ctx context.Context, chainName string) error {
    client, err := idx.manager.GetClient(chainName)
    if err != nil {
        return fmt.Errorf("failed to get client: %w", err)
    }
    
    // Get starting block
    startBlock, err := idx.getStartBlock(ctx, chainName, client)
    if err != nil {
        return fmt.Errorf("failed to get start block: %w", err)
    }
    
    fmt.Printf("Starting indexer for %s from block %d\n", chainName, startBlock)
    
    // Start real-time listener if WebSocket available
    if client.IsWebSocketAvailable() {
        go idx.listenToRealtimeEvents(ctx, chainName, client)
    }
    
    // Start historical indexing
    return idx.indexHistoricalEvents(ctx, chainName, client, startBlock)
}

func (idx *Indexer) getStartBlock(ctx context.Context, chain string, client *blockchain.Client) (uint64, error) {
    // Check last indexed block from storage
    lastIndexed, err := idx.storage.GetLastIndexedBlock(ctx, chain)
    if err == nil && lastIndexed > 0 {
        return lastIndexed + 1, nil
    }
    
    // If no previous index or error, start from current - config.StartFromBlock
    currentBlock, err := client.GetBlockNumber(ctx)
    if err != nil {
        return 0, err
    }
    
    startFrom := idx.config.StartFromBlock
    if startFrom == 0 {
        startFrom = 1000 // Default to last 1000 blocks
    }
    
    if currentBlock > startFrom {
        return currentBlock - startFrom, nil
    }
    
    return 0, nil
}

func (idx *Indexer) indexHistoricalEvents(ctx context.Context, chain string, client *blockchain.Client, fromBlock uint64) error {
    currentBlock, err := client.GetBlockNumber(ctx)
    if err != nil {
        return err
    }
    
    // Account for confirmations
    if currentBlock > idx.config.BlockConfirmations {
        currentBlock -= idx.config.BlockConfirmations
    }
    
    batchSize := idx.config.BatchSize
    workerChan := make(chan blockRange, idx.config.WorkerCount)
    
    // Start workers
    workerCtx, cancel := context.WithCancel(ctx)
    defer cancel()
    
    for i := 0; i < idx.config.WorkerCount; i++ {
        idx.wg.Add(1)
        go idx.worker(workerCtx, chain, client, workerChan)
    }
    
    // Send work to workers
    for fromBlock < currentBlock {
        toBlock := fromBlock + batchSize - 1
        if toBlock > currentBlock {
            toBlock = currentBlock
        }
        
        select {
        case workerChan <- blockRange{from: fromBlock, to: toBlock}:
            fromBlock = toBlock + 1
        case <-ctx.Done():
            close(workerChan)
            return ctx.Err()
        }
    }
    
    close(workerChan)
    return nil
}

type blockRange struct {
    from uint64
    to   uint64
}

func (idx *Indexer) worker(ctx context.Context, chain string, client *blockchain.Client, jobs <-chan blockRange) {
    defer idx.wg.Done()
    
    for job := range jobs {
        if err := idx.processBatch(ctx, chain, client, job.from, job.to); err != nil {
            fmt.Printf("Error processing batch %d-%d: %v\n", job.from, job.to, err)
            // Implement retry logic here if needed
        }
        
        // Small delay to avoid overwhelming the RPC
        time.Sleep(100 * time.Millisecond)
    }
}

func (idx *Indexer) processBatch(ctx context.Context, chain string, client *blockchain.Client, from, to uint64) error {
    signatures := idx.getAllEventSignatures()
    if len(signatures) == 0 {
        return nil // No events to index
    }
    
    filter := blockchain.EventFilter{
        FromBlock: big.NewInt(int64(from)),
        ToBlock:   big.NewInt(int64(to)),
        Topics:    [][]common.Hash{signatures},
    }
    
    logs, err := client.GetLogs(ctx, filter)
    if err != nil {
        return fmt.Errorf("failed to get logs: %w", err)
    }
    
    events := make([]*Event, 0, len(logs))
    for _, log := range logs {
        event := idx.processLog(chain, log)
        if event != nil {
            events = append(events, event)
        }
    }
    
    // Save events
    if len(events) > 0 {
        if err := idx.storage.SaveEvents(ctx, events); err != nil {
            return fmt.Errorf("failed to save events: %w", err)
        }
    }
    
    // Update last indexed block
    block := &Block{
        Chain:     chain,
        Number:    to,
        Timestamp: time.Now(),
    }
    
    if err := idx.storage.SaveBlock(ctx, block); err != nil {
        return fmt.Errorf("failed to save block: %w", err)
    }
    
    fmt.Printf("[%s] Indexed blocks %d-%d: %d events\n", chain, from, to, len(events))
    return nil
}

func (idx *Indexer) processLog(chain string, log types.Log) *Event {
    idx.mu.RLock()
    defer idx.mu.RUnlock()
    
    for _, processor := range idx.processors {
        signatures := processor.GetEventSignatures()
        for _, sig := range signatures {
            if len(log.Topics) > 0 && log.Topics[0] == sig {
                event, err := processor.Process(log)
                if err != nil {
                    fmt.Printf("Error processing log: %v\n", err)
                    return nil
                }
                if event != nil {
                    event.Chain = chain
                    return event
                }
            }
        }
    }
    
    return nil
}

func (idx *Indexer) getAllEventSignatures() []common.Hash {
    idx.mu.RLock()
    defer idx.mu.RUnlock()
    
    sigMap := make(map[common.Hash]bool)
    for _, processor := range idx.processors {
        for _, sig := range processor.GetEventSignatures() {
            sigMap[sig] = true
        }
    }
    
    signatures := make([]common.Hash, 0, len(sigMap))
    for sig := range sigMap {
        signatures = append(signatures, sig)
    }
    
    return signatures
}

func (idx *Indexer) listenToRealtimeEvents(ctx context.Context, chain string, client *blockchain.Client) {
    fmt.Printf("Starting real-time event listener for %s\n", chain)
    
    signatures := idx.getAllEventSignatures()
    if len(signatures) == 0 {
        return
    }
    
    filter := blockchain.EventFilter{
        Topics: [][]common.Hash{signatures},
    }
    
    logs, sub, err := client.SubscribeLogs(ctx, filter)
    if err != nil {
        fmt.Printf("Failed to subscribe to logs for %s: %v\n", chain, err)
        return
    }
    
    for {
        select {
        case err := <-sub.Err():
            fmt.Printf("Subscription error for %s: %v\n", chain, err)
            // Implement reconnection logic here
            return
            
        case log := <-logs:
            event := idx.processLog(chain, log)
            if event != nil {
                if err := idx.storage.SaveEvents(ctx, []*Event{event}); err != nil {
                    fmt.Printf("Failed to save real-time event: %v\n", err)
                }
                fmt.Printf("[%s] New event: %s at block %d\n", chain, event.EventName, log.BlockNumber)
            }
            
        case <-ctx.Done():
            sub.Unsubscribe()
            return
        }
    }
}

func (idx *Indexer) Stop() {
    fmt.Println("Stopping indexer...")
    idx.wg.Wait()
    fmt.Println("Indexer stopped")
}