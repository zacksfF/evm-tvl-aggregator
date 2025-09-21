package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Example data structures for expected results
type ChainTVL struct {
	ChainID     int       `json:"chain_id"`
	ChainName   string    `json:"chain_name"`
	TVL         float64   `json:"tvl"`
	TVLChange24h float64  `json:"tvl_change_24h"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProtocolTVL struct {
	ProtocolID   string              `json:"protocol_id"`
	Name         string              `json:"name"`
	Category     string              `json:"category"`
	TotalTVL     float64             `json:"total_tvl"`
	ChainTVLs    map[string]float64  `json:"chain_tvls"`
	TVLHistory   []TVLSnapshot       `json:"tvl_history,omitempty"`
	UpdatedAt    time.Time           `json:"updated_at"`
}

type TVLSnapshot struct {
	Timestamp time.Time `json:"timestamp"`
	TVL       float64   `json:"tvl"`
}

type AggregatedTVL struct {
	TotalTVL          float64            `json:"total_tvl"`
	TotalProtocols    int                `json:"total_protocols"`
	ChainBreakdown    []ChainTVL         `json:"chain_breakdown"`
	TopProtocols      []ProtocolTVL      `json:"top_protocols"`
	CategoryBreakdown map[string]float64 `json:"category_breakdown"`
	LastUpdated       time.Time          `json:"last_updated"`
}

type TokenDistribution struct {
	Token      string  `json:"token"`
	Amount     float64 `json:"amount"`
	ValueUSD   float64 `json:"value_usd"`
	Percentage float64 `json:"percentage"`
}

type ProtocolDetails struct {
	Protocol          ProtocolTVL         `json:"protocol"`
	TokenDistribution []TokenDistribution `json:"token_distribution"`
	PoolsOrVaults     []PoolInfo          `json:"pools_or_vaults"`
	RiskMetrics       RiskMetrics         `json:"risk_metrics"`
}

type PoolInfo struct {
	PoolID      string  `json:"pool_id"`
	Name        string  `json:"name"`
	TVL         float64 `json:"tvl"`
	APY         float64 `json:"apy"`
	Volume24h   float64 `json:"volume_24h"`
}

type RiskMetrics struct {
	AuditScore     float64 `json:"audit_score"`
	TimeActive     int     `json:"time_active_days"`
	ILProtection   bool    `json:"il_protection"`
	MultisigSecured bool   `json:"multisig_secured"`
}

func main() {
	fmt.Println("=== EVM TVL Aggregator - Example Expected Results ===\n")

	// Example 1: Aggregated TVL across all chains
	aggregatedTVL := AggregatedTVL{
		TotalTVL:       125750500000.50, // $125.75B
		TotalProtocols: 234,
		ChainBreakdown: []ChainTVL{
			{ChainID: 1, ChainName: "Ethereum", TVL: 85500000000.00, TVLChange24h: 2.5, UpdatedAt: time.Now()},
			{ChainID: 56, ChainName: "BSC", TVL: 15750500000.00, TVLChange24h: -1.2, UpdatedAt: time.Now()},
			{ChainID: 137, ChainName: "Polygon", TVL: 8500000000.00, TVLChange24h: 3.8, UpdatedAt: time.Now()},
			{ChainID: 42161, ChainName: "Arbitrum", TVL: 7250000000.00, TVLChange24h: 5.2, UpdatedAt: time.Now()},
			{ChainID: 10, ChainName: "Optimism", TVL: 4750000000.50, TVLChange24h: 1.7, UpdatedAt: time.Now()},
			{ChainID: 43114, ChainName: "Avalanche", TVL: 4000000000.00, TVLChange24h: -0.5, UpdatedAt: time.Now()},
		},
		TopProtocols: []ProtocolTVL{
			{
				ProtocolID: "aave-v3",
				Name:       "Aave V3",
				Category:   "Lending",
				TotalTVL:   28500000000.00,
				ChainTVLs: map[string]float64{
					"ethereum": 18500000000.00,
					"polygon":  4000000000.00,
					"arbitrum": 3500000000.00,
					"optimism": 2500000000.00,
				},
				UpdatedAt: time.Now(),
			},
			{
				ProtocolID: "uniswap-v3",
				Name:       "Uniswap V3",
				Category:   "DEX",
				TotalTVL:   22750000000.00,
				ChainTVLs: map[string]float64{
					"ethereum": 15000000000.00,
					"polygon":  3750000000.00,
					"arbitrum": 2500000000.00,
					"optimism": 1500000000.00,
				},
				UpdatedAt: time.Now(),
			},
		},
		CategoryBreakdown: map[string]float64{
			"Lending":       45500000000.00,
			"DEX":           38750000000.00,
			"Yield":         25500000000.00,
			"Derivatives":   8500000000.00,
			"Staking":       7500500000.50,
		},
		LastUpdated: time.Now(),
	}

	// Example 2: Detailed protocol information
	protocolDetails := ProtocolDetails{
		Protocol: ProtocolTVL{
			ProtocolID: "compound-v3",
			Name:       "Compound V3",
			Category:   "Lending",
			TotalTVL:   8750000000.00,
			ChainTVLs: map[string]float64{
				"ethereum": 6500000000.00,
				"polygon":  1250000000.00,
				"arbitrum": 1000000000.00,
			},
			TVLHistory: []TVLSnapshot{
				{Timestamp: time.Now().Add(-24 * time.Hour), TVL: 8500000000.00},
				{Timestamp: time.Now().Add(-12 * time.Hour), TVL: 8650000000.00},
				{Timestamp: time.Now(), TVL: 8750000000.00},
			},
			UpdatedAt: time.Now(),
		},
		TokenDistribution: []TokenDistribution{
			{Token: "USDC", Amount: 3500000000, ValueUSD: 3500000000, Percentage: 40.0},
			{Token: "ETH", Amount: 1250000, ValueUSD: 2750000000, Percentage: 31.4},
			{Token: "WBTC", Amount: 45000, ValueUSD: 1500000000, Percentage: 17.1},
			{Token: "DAI", Amount: 1000000000, ValueUSD: 1000000000, Percentage: 11.5},
		},
		PoolsOrVaults: []PoolInfo{
			{PoolID: "usdc-market", Name: "USDC Market", TVL: 3500000000, APY: 4.5, Volume24h: 125000000},
			{PoolID: "eth-market", Name: "ETH Market", TVL: 2750000000, APY: 2.8, Volume24h: 95000000},
			{PoolID: "wbtc-market", Name: "WBTC Market", TVL: 1500000000, APY: 1.9, Volume24h: 45000000},
		},
		RiskMetrics: RiskMetrics{
			AuditScore:      9.2,
			TimeActive:      1095, // 3 years
			ILProtection:    false,
			MultisigSecured: true,
		},
	}

	// Print Example 1: Aggregated TVL
	fmt.Println("1. AGGREGATED TVL RESPONSE:")
	fmt.Println("----------------------------")
	printJSON(aggregatedTVL)

	// Print Example 2: Protocol Details
	fmt.Println("\n2. PROTOCOL DETAILS RESPONSE:")
	fmt.Println("-------------------------------")
	printJSON(protocolDetails)

	// Example 3: Real-time TVL update event
	fmt.Println("\n3. REAL-TIME TVL UPDATE EVENT:")
	fmt.Println("---------------------------------")
	realtimeUpdate := map[string]interface{}{
		"event_type": "tvl_update",
		"protocol_id": "aave-v3",
		"chain": "ethereum",
		"old_tvl": 18450000000.00,
		"new_tvl": 18500000000.00,
		"change_percentage": 0.27,
		"timestamp": time.Now(),
		"block_number": 19234567,
		"tx_hash": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bA45a3b2f6c92a3b4a3b8b0e",
	}
	printJSON(realtimeUpdate)

	// Example 4: API Response for chains endpoint
	fmt.Println("\n4. /api/chains ENDPOINT RESPONSE:")
	fmt.Println("-----------------------------------")
	chainsResponse := map[string]interface{}{
		"success": true,
		"data": []map[string]interface{}{
			{
				"chain_id": 1,
				"name": "Ethereum",
				"tvl": 85500000000.00,
				"market_share": 68.0,
				"protocols_count": 156,
				"daily_volume": 2500000000.00,
				"gas_price_gwei": 25.5,
			},
			{
				"chain_id": 56,
				"name": "BSC",
				"tvl": 15750500000.00,
				"market_share": 12.5,
				"protocols_count": 89,
				"daily_volume": 850000000.00,
				"gas_price_gwei": 3.2,
			},
		},
		"timestamp": time.Now(),
	}
	printJSON(chainsResponse)

	// Example 5: Historical TVL data
	fmt.Println("\n5. HISTORICAL TVL DATA:")
	fmt.Println("------------------------")
	historicalData := map[string]interface{}{
		"protocol": "uniswap-v3",
		"chain": "ethereum",
		"timeframe": "7d",
		"data_points": []map[string]interface{}{
			{"timestamp": "2024-01-15T00:00:00Z", "tvl": 14500000000.00, "volume": 1250000000.00},
			{"timestamp": "2024-01-16T00:00:00Z", "tvl": 14650000000.00, "volume": 1380000000.00},
			{"timestamp": "2024-01-17T00:00:00Z", "tvl": 14800000000.00, "volume": 1425000000.00},
			{"timestamp": "2024-01-18T00:00:00Z", "tvl": 14750000000.00, "volume": 1390000000.00},
			{"timestamp": "2024-01-19T00:00:00Z", "tvl": 14900000000.00, "volume": 1510000000.00},
			{"timestamp": "2024-01-20T00:00:00Z", "tvl": 14950000000.00, "volume": 1485000000.00},
			{"timestamp": "2024-01-21T00:00:00Z", "tvl": 15000000000.00, "volume": 1550000000.00},
		},
		"average_tvl": 14792857142.86,
		"tvl_change": 3.45,
		"highest_tvl": 15000000000.00,
		"lowest_tvl": 14500000000.00,
	}
	printJSON(historicalData)
}

func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
}