package models

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Event represents a blockchain event
type Event struct {
	ID              uint64         `json:"id" db:"id"`
	Chain           string         `json:"chain" db:"chain"`
	Protocol        string         `json:"protocol" db:"protocol"`
	BlockNumber     uint64         `json:"block_number" db:"block_number"`
	BlockHash       string         `json:"block_hash" db:"block_hash"`
	TransactionHash string         `json:"transaction_hash" db:"transaction_hash"`
	LogIndex        uint           `json:"log_index" db:"log_index"`
	Address         common.Address `json:"address" db:"address"`
	EventName       string         `json:"event_name" db:"event_name"`
	EventSignature  string         `json:"event_signature" db:"event_signature"`
	Data            EventData      `json:"data" db:"data"`
	Timestamp       time.Time      `json:"timestamp" db:"timestamp"`
}

// EventData is a flexible map for event data
type EventData map[string]interface{}

// Block represents an indexed block
type Block struct {
	ID        uint64    `json:"id" db:"id"`
	Chain     string    `json:"chain" db:"chain"`
	Number    uint64    `json:"number" db:"number"`
	Hash      string    `json:"hash" db:"hash"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}
