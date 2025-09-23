package models

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// TVLSnapshot represents a point-in-time TVL measurement
type TVLSnapshot struct {
	ID          uint64               `json:"id" db:"id"`
	ProtocolID  uint64               `json:"protocol_id" db:"protocol_id"`
	Protocol    string               `json:"protocol" db:"protocol"`
	Chain       string               `json:"chain" db:"chain"`
	BlockNumber uint64               `json:"block_number" db:"block_number"`
	TotalUSD    *big.Float           `json:"total_usd"`
	Breakdown   map[string]*AssetTVL `json:"breakdown"`
	Timestamp   time.Time            `json:"timestamp" db:"timestamp"`
}

// TVLData represents current TVL data
type TVLData struct {
	Protocol  string               `json:"protocol"`
	Timestamp time.Time            `json:"timestamp"`
	Chains    map[string]*ChainTVL `json:"chains"`
	TotalUSD  *big.Float           `json:"total_usd"`
}

// ChainTVL represents TVL for a specific chain
type ChainTVL struct {
	Chain       string      `json:"chain"`
	BlockNumber uint64      `json:"block_number"`
	Assets      []*AssetTVL `json:"assets"`
	TotalUSD    *big.Float  `json:"total_usd"`
}

// AssetTVL represents TVL for a specific asset
type AssetTVL struct {
	Token      common.Address `json:"token"`
	Symbol     string         `json:"symbol"`
	Name       string         `json:"name,omitempty"`
	Amount     *big.Int       `json:"amount"`
	Decimals   uint8          `json:"decimals"`
	PriceUSD   *big.Float     `json:"price_usd"`
	ValueUSD   *big.Float     `json:"value_usd"`
	Percentage float64        `json:"percentage,omitempty"`
}

// TVLHistory represents historical TVL data
type TVLHistory struct {
	Protocol   string         `json:"protocol"`
	Chain      string         `json:"chain,omitempty"`
	Period     string         `json:"period"`
	DataPoints []TVLDataPoint `json:"data_points"`
}

// TVLDataPoint represents a single TVL data point
type TVLDataPoint struct {
	Timestamp   time.Time  `json:"timestamp"`
	BlockNumber uint64     `json:"block_number,omitempty"`
	TVL         *big.Float `json:"tvl"`
	Change24h   float64    `json:"change_24h,omitempty"`
}

// AggregatedTVL represents TVL across all protocols
type AggregatedTVL struct {
	Timestamp    time.Time             `json:"timestamp"`
	TotalUSD     *big.Float            `json:"total_usd"`
	Protocols    map[string]*TVLData   `json:"protocols"`
	ChainTotals  map[string]*big.Float `json:"chain_totals"`
	TopProtocols []ProtocolRanking     `json:"top_protocols,omitempty"`
	Change24h    float64               `json:"change_24h,omitempty"`
}

// ProtocolRanking for leaderboards
type ProtocolRanking struct {
	Rank        int        `json:"rank"`
	Protocol    string     `json:"protocol"`
	TVL         *big.Float `json:"tvl"`
	MarketShare float64    `json:"market_share"`
	Change24h   float64    `json:"change_24h"`
}
