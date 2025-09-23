package models

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Token represents an ERC20 token
type Token struct {
	ID          uint64         `json:"id" db:"id"`
	Address     common.Address `json:"address" db:"address"`
	Chain       string         `json:"chain" db:"chain"`
	Symbol      string         `json:"symbol" db:"symbol"`
	Name        string         `json:"name" db:"name"`
	Decimals    uint8          `json:"decimals" db:"decimals"`
	TotalSupply *big.Int       `json:"total_supply" db:"total_supply"`
	Logo        string         `json:"logo" db:"logo"`
	PriceUSD    *big.Float     `json:"price_usd"`
	MarketCap   *big.Float     `json:"market_cap"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

// TokenPrice represents token price data
type TokenPrice struct {
	ID         uint64         `json:"id" db:"id"`
	TokenID    uint64         `json:"token_id" db:"token_id"`
	Address    common.Address `json:"address" db:"address"`
	Symbol     string         `json:"symbol" db:"symbol"`
	PriceUSD   *big.Float     `json:"price_usd" db:"price_usd"`
	Source     string         `json:"source" db:"source"`
	Confidence float64        `json:"confidence" db:"confidence"`
	Timestamp  time.Time      `json:"timestamp" db:"timestamp"`
}

// PriceSource represents where prices come from
type PriceSource string

const (
	PriceSourceCoinGecko PriceSource = "coingecko"
	PriceSourceChainlink PriceSource = "chainlink"
	PriceSourceUniswap   PriceSource = "uniswap"
	PriceSourceOracle    PriceSource = "oracle"
	PriceSourceManual    PriceSource = "manual"
)
