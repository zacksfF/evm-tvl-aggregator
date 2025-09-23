// internal/indexer/storage.go
package indexer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
    db *sql.DB
}

func NewPostgresStorage(connectionString string) (*PostgresStorage, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    storage := &PostgresStorage{db: db}
    if err := storage.createTables(); err != nil {
        return nil, err
    }
    
    return storage, nil
}

func (s *PostgresStorage) createTables() error {
    queries := []string{
        `CREATE TABLE IF NOT EXISTS indexed_blocks (
            id SERIAL PRIMARY KEY,
            chain VARCHAR(50) NOT NULL,
            block_number BIGINT NOT NULL,
            block_hash VARCHAR(66),
            indexed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(chain, block_number)
        )`,
        
        `CREATE TABLE IF NOT EXISTS events (
            id SERIAL PRIMARY KEY,
            chain VARCHAR(50) NOT NULL,
            block_number BIGINT NOT NULL,
            block_hash VARCHAR(66),
            transaction_hash VARCHAR(66) NOT NULL,
            log_index INT NOT NULL,
            address VARCHAR(42) NOT NULL,
            event_name VARCHAR(100) NOT NULL,
            protocol VARCHAR(100),
            data JSONB,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            UNIQUE(chain, transaction_hash, log_index)
        )`,
        
        `CREATE INDEX IF NOT EXISTS idx_events_chain_block ON events(chain, block_number)`,
        `CREATE INDEX IF NOT EXISTS idx_events_protocol ON events(protocol)`,
        `CREATE INDEX IF NOT EXISTS idx_events_address ON events(address)`,
        `CREATE INDEX IF NOT EXISTS idx_events_name ON events(event_name)`,
    }
    
    for _, query := range queries {
        if _, err := s.db.Exec(query); err != nil {
            return fmt.Errorf("failed to execute query: %w", err)
        }
    }
    
    return nil
}

func (s *PostgresStorage) SaveBlock(ctx context.Context, block *Block) error {
    query := `
        INSERT INTO indexed_blocks (chain, block_number, block_hash, indexed_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (chain, block_number) 
        DO UPDATE SET indexed_at = EXCLUDED.indexed_at
    `
    
    _, err := s.db.ExecContext(ctx, query, 
        block.Chain, 
        block.Number, 
        block.Hash, 
        block.Timestamp,
    )
    
    return err
}

func (s *PostgresStorage) GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error) {
    var blockNumber uint64
    query := `
        SELECT block_number 
        FROM indexed_blocks 
        WHERE chain = $1 
        ORDER BY block_number DESC 
        LIMIT 1
    `
    
    err := s.db.QueryRowContext(ctx, query, chain).Scan(&blockNumber)
    if err == sql.ErrNoRows {
        return 0, nil
    }
    
    return blockNumber, err
}

func (s *PostgresStorage) SaveEvents(ctx context.Context, events []*Event) error {
    if len(events) == 0 {
        return nil
    }
    
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    stmt, err := tx.PrepareContext(ctx, `
        INSERT INTO events (
            chain, block_number, block_hash, transaction_hash, 
            log_index, address, event_name, protocol, data
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
        ON CONFLICT (chain, transaction_hash, log_index) DO NOTHING
    `)
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    for _, event := range events {
        dataJSON, err := json.Marshal(event.Data)
        if err != nil {
            return err
        }
        
        _, err = stmt.ExecContext(ctx,
            event.Chain,
            event.BlockNumber,
            event.BlockHash,
            event.TransactionHash,
            event.LogIndex,
            event.Address.Hex(),
            event.EventName,
            event.Protocol,
            dataJSON,
        )
        if err != nil {
            return err
        }
    }
    
    return tx.Commit()
}

func (s *PostgresStorage) GetEvents(ctx context.Context, chain string, from, to uint64) ([]*Event, error) {
    query := `
        SELECT 
            id, chain, block_number, block_hash, transaction_hash,
            log_index, address, event_name, protocol, data, created_at
        FROM events
        WHERE chain = $1 AND block_number >= $2 AND block_number <= $3
        ORDER BY block_number, log_index
    `
    
    rows, err := s.db.QueryContext(ctx, query, chain, from, to)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var events []*Event
    for rows.Next() {
        var event Event
        var address string
        var dataJSON []byte
        
        err := rows.Scan(
            &event.ID,
            &event.Chain,
            &event.BlockNumber,
            &event.BlockHash,
            &event.TransactionHash,
            &event.LogIndex,
            &address,
            &event.EventName,
            &event.Protocol,
            &dataJSON,
            &event.Timestamp,
        )
        if err != nil {
            return nil, err
        }
        
        event.Address = common.HexToAddress(address)
        if err := json.Unmarshal(dataJSON, &event.Data); err != nil {
            return nil, err
        }
        
        events = append(events, &event)
    }
    
    return events, nil
}

func (s *PostgresStorage) GetIndexerStats(ctx context.Context, chain string) (*IndexerStats, error) {
    stats := &IndexerStats{
        Chain:          chain,
        LastUpdateTime: time.Now(),
    }
    
    // Get last indexed block
    lastBlock, _ := s.GetLastIndexedBlock(ctx, chain)
    stats.LastIndexedBlock = lastBlock
    
    // Get event count
    var count int64
    err := s.db.QueryRowContext(ctx, 
        "SELECT COUNT(*) FROM events WHERE chain = $1", 
        chain,
    ).Scan(&count)
    if err != nil {
        return nil, err
    }
    stats.EventsCount = uint64(count)
    
    return stats, nil
}

func (s *PostgresStorage) Close() error {
    return s.db.Close()
}

// MemoryStorage implements Storage interface with in-memory storage
type MemoryStorage struct {
    mu          sync.RWMutex
    blocks      map[string]map[uint64]*Block  // chain -> block_number -> block
    events      map[string][]*Event           // chain -> events
    lastBlocks  map[string]uint64             // chain -> last_block_number
}

func NewMemoryStorage() *MemoryStorage {
    return &MemoryStorage{
        blocks:     make(map[string]map[uint64]*Block),
        events:     make(map[string][]*Event),
        lastBlocks: make(map[string]uint64),
    }
}

func (s *MemoryStorage) SaveBlock(ctx context.Context, block *Block) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if s.blocks[block.Chain] == nil {
        s.blocks[block.Chain] = make(map[uint64]*Block)
    }
    
    s.blocks[block.Chain][block.Number] = block
    
    // Update last block number
    if block.Number > s.lastBlocks[block.Chain] {
        s.lastBlocks[block.Chain] = block.Number
    }
    
    return nil
}

func (s *MemoryStorage) GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    return s.lastBlocks[chain], nil
}

func (s *MemoryStorage) SaveEvents(ctx context.Context, events []*Event) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    for _, event := range events {
        s.events[event.Chain] = append(s.events[event.Chain], event)
    }
    
    return nil
}

func (s *MemoryStorage) GetEvents(ctx context.Context, chain string, from, to uint64) ([]*Event, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    var filtered []*Event
    for _, event := range s.events[chain] {
        if event.BlockNumber >= from && event.BlockNumber <= to {
            filtered = append(filtered, event)
        }
    }
    
    return filtered, nil
}

func (s *MemoryStorage) GetIndexerStats(ctx context.Context, chain string) (*IndexerStats, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    stats := &IndexerStats{
        Chain:             chain,
        LastIndexedBlock:  s.lastBlocks[chain],
        EventsCount:       uint64(len(s.events[chain])),
        LastUpdateTime:    time.Now(),
    }
    
    return stats, nil
}