package models

import (
	"math/big"
	"time"
)

// Chain represents a blockchain
type Chain struct {
	ID          uint64    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	ChainID     *big.Int  `json:"chain_id" db:"chain_id"`
	RPCEndpoint string    `json:"rpc_endpoint" db:"rpc_endpoint"`
	WSEndpoint  string    `json:"ws_endpoint" db:"ws_endpoint"`
	Explorer    string    `json:"explorer" db:"explorer"`
	NativeToken string    `json:"native_token" db:"native_token"`
	BlockTime   int       `json:"block_time" db:"block_time"` // seconds
	IsTestnet   bool      `json:"is_testnet" db:"is_testnet"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// ChainStats represents statistics for a chain
type ChainStats struct {
	Chain            string     `json:"chain"`
	LastIndexedBlock uint64     `json:"last_indexed_block"`
	CurrentBlock     uint64     `json:"current_block"`
	TotalProtocols   int        `json:"total_protocols"`
	TotalTVL         *big.Float `json:"total_tvl"`
	LastUpdateTime   time.Time  `json:"last_update_time"`
}
