package models

import "time"

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// TVLResponse represents the total TVL response
type TVLResponse struct {
	TotalTVL  float64                    `json:"total_tvl"`
	Timestamp time.Time                  `json:"timestamp"`
	Protocols map[string]ProtocolTVLData `json:"protocols"`
	Chains    map[string]float64         `json:"chains"`
}

// ProtocolTVLResponse represents TVL data for a specific protocol
type ProtocolTVLResponse struct {
	Protocol  string                     `json:"protocol"`
	TotalTVL  float64                    `json:"total_tvl"`
	Timestamp time.Time                  `json:"timestamp"`
	Chains    map[string]ChainTVLData    `json:"chains"`
}

// ProtocolTVLData represents basic protocol TVL data
type ProtocolTVLData struct {
	TotalUSD  float64   `json:"total_usd"`
	Timestamp time.Time `json:"timestamp"`
}

// ChainTVLData represents TVL data for a specific chain
type ChainTVLData struct {
	TotalUSD float64     `json:"total_usd"`
	Assets   []AssetData `json:"assets"`
}

// AssetData represents asset information
type AssetData struct {
	Token     string  `json:"token"`
	Symbol    string  `json:"symbol"`
	Amount    string  `json:"amount"`
	Decimals  int     `json:"decimals"`
	ValueUSD  float64 `json:"value_usd"`
}

// ProtocolsResponse represents the protocols list response
type ProtocolsResponse struct {
	Protocols []ProtocolInfo `json:"protocols"`
	Count     int            `json:"count"`
}

// ProtocolInfo represents basic protocol information
type ProtocolInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Chains      []string `json:"chains"`
}

// ChainsResponse represents the chains list response
type ChainsResponse struct {
	Chains []string `json:"chains"`
	Count  int      `json:"count"`
}

// StatsResponse represents system statistics
type StatsResponse struct {
	TotalTVL      float64   `json:"total_tvl"`
	ProtocolCount int       `json:"protocol_count"`
	ChainCount    int       `json:"chain_count"`
	LastUpdated   time.Time `json:"last_updated"`
}

// HistoricalTVLResponse represents historical TVL data
type HistoricalTVLResponse struct {
	Protocol string              `json:"protocol"`
	Chain    string              `json:"chain"`
	Period   string              `json:"period"`
	History  []HistoricalTVLData `json:"history"`
}

// HistoricalTVLData represents a single historical data point
type HistoricalTVLData struct {
	Timestamp time.Time `json:"timestamp"`
	TVL       float64   `json:"tvl"`
}

// UIState represents the current state of the UI
type UIState struct {
	CurrentView    string
	SelectedIndex  int
	Loading        bool
	Error          error
	LastUpdate     time.Time
	RefreshEnabled bool
}

// ChartData represents data for rendering charts
type ChartData struct {
	Title   string
	Labels  []string
	Values  []float64
	Colors  []string
	MaxY    float64
	MinY    float64
}

// TableData represents data for rendering tables
type TableData struct {
	Headers []string
	Rows    [][]string
	Width   int
	Height  int
}