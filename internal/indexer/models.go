// internal/indexer/models.go
package indexer

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Block represents an indexed block
type Block struct {
	ID        uint64    `json:"id"`
	Chain     string    `json:"chain"`
	Number    uint64    `json:"number"`
	Hash      string    `json:"hash,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Event represents a decoded blockchain event
type Event struct {
	ID              uint64         `json:"id"`
	Chain           string         `json:"chain"`
	BlockNumber     uint64         `json:"block_number"`
	BlockHash       string         `json:"block_hash"`
	TransactionHash string         `json:"transaction_hash"`
	LogIndex        uint           `json:"log_index"`
	Address         common.Address `json:"address"`
	EventName       string         `json:"event_name"`
	Protocol        string         `json:"protocol"`
	Data            EventData      `json:"data"`
	Timestamp       time.Time      `json:"timestamp"`
}

// EventData contains decoded event data
type EventData map[string]interface{}

// Common event types
type TransferEventData struct {
	From   common.Address `json:"from"`
	To     common.Address `json:"to"`
	Amount *big.Int       `json:"amount"`
}

type DepositEventData struct {
	User   common.Address `json:"user"`
	Amount *big.Int       `json:"amount"`
	Asset  common.Address `json:"asset,omitempty"`
}

type WithdrawEventData struct {
	User   common.Address `json:"user"`
	Amount *big.Int       `json:"amount"`
	Asset  common.Address `json:"asset,omitempty"`
}

type SwapEventData struct {
	User      common.Address `json:"user"`
	TokenIn   common.Address `json:"token_in"`
	TokenOut  common.Address `json:"token_out"`
	AmountIn  *big.Int       `json:"amount_in"`
	AmountOut *big.Int       `json:"amount_out"`
}

type LiquidityEventData struct {
	User      common.Address `json:"user"`
	Token0    common.Address `json:"token0"`
	Token1    common.Address `json:"token1"`
	Amount0   *big.Int       `json:"amount0"`
	Amount1   *big.Int       `json:"amount1"`
	Liquidity *big.Int       `json:"liquidity"`
}

// IndexerStats provides statistics about indexing progress
type IndexerStats struct {
	Chain            string    `json:"chain"`
	LastIndexedBlock uint64    `json:"last_indexed_block"`
	CurrentBlock     uint64    `json:"current_block"`
	EventsCount      uint64    `json:"events_count"`
	IndexingProgress float64   `json:"indexing_progress"`
	LastUpdateTime   time.Time `json:"last_update_time"`
}
