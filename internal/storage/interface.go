package storage

import (
	"context"
	"time"

	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
)

// Storage defines the main storage interface
type Storage interface {
	ProtocolStorage
	TVLStorage
	ChainStorage
	EventStorage
	TokenStorage

	// Transaction support
	BeginTx(ctx context.Context) (Tx, error)
	Close() error
}

// Tx represents a database transaction
type Tx interface {
	Storage
	Commit() error
	Rollback() error
}

// ProtocolStorage handles protocol data
type ProtocolStorage interface {
	GetProtocol(ctx context.Context, name string) (*models.Protocol, error)
	GetProtocols(ctx context.Context) ([]*models.Protocol, error)
	SaveProtocol(ctx context.Context, protocol *models.Protocol) error
	UpdateProtocol(ctx context.Context, protocol *models.Protocol) error
	DeleteProtocol(ctx context.Context, name string) error
}

// TVLStorage handles TVL data
type TVLStorage interface {
	SaveTVLSnapshot(ctx context.Context, snapshot *models.TVLSnapshot) error
	GetLatestTVL(ctx context.Context, protocol, chain string) (*models.TVLSnapshot, error)
	GetHistoricalTVL(ctx context.Context, protocol, chain string, from, to time.Time) ([]*models.TVLSnapshot, error)
	GetTVLByBlock(ctx context.Context, protocol, chain string, blockNumber uint64) (*models.TVLSnapshot, error)
	GetAggregatedTVL(ctx context.Context) (*models.AggregatedTVL, error)
}

// ChainStorage handles chain data
type ChainStorage interface {
	GetChain(ctx context.Context, name string) (*models.Chain, error)
	GetChains(ctx context.Context) ([]*models.Chain, error)
	SaveChain(ctx context.Context, chain *models.Chain) error
	UpdateChainStats(ctx context.Context, stats *models.ChainStats) error
}

// EventStorage handles event data
type EventStorage interface {
	SaveEvents(ctx context.Context, events []*models.Event) error
	GetEvents(ctx context.Context, filter EventFilter) ([]*models.Event, error)
	SaveBlock(ctx context.Context, block *models.Block) error
	GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error)
}

// TokenStorage handles token data
type TokenStorage interface {
	GetToken(ctx context.Context, address, chain string) (*models.Token, error)
	SaveToken(ctx context.Context, token *models.Token) error
	UpdateTokenPrice(ctx context.Context, price *models.TokenPrice) error
	GetTokenPrices(ctx context.Context, addresses []string) (map[string]*models.TokenPrice, error)
}

// EventFilter for querying events
type EventFilter struct {
	Chain     string
	Protocol  string
	FromBlock uint64
	ToBlock   uint64
	EventName string
	Address   string
	Limit     int
	Offset    int
}
