// internal/api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/zacksfF/evm-tvl-aggregator/internal/aggregator"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage"
)

type Handler struct {
	calculator *aggregator.TVLCalculator
	storage    storage.Storage
	cache      Cache
}

type Cache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, ttl time.Duration) error
}

func NewHandler(calculator *aggregator.TVLCalculator, storage storage.Storage) *Handler {
	return &Handler{
		calculator: calculator,
		storage:    storage,
	}
}

// GET /api/v1/tvl
func (h *Handler) GetTotalTVL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Check cache
	cacheKey := "tvl:total"
	if h.cache != nil {
		if cached, err := h.cache.Get(cacheKey); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached)
			return
		}
	}

	// Calculate aggregated TVL
	aggregatedTVL, err := h.calculator.CalculateAllProtocolsTVL(ctx)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert to JSON-friendly format
	totalFloat, _ := aggregatedTVL.TotalUSD.Float64()

	response := map[string]interface{}{
		"total_tvl": totalFloat,
		"timestamp": aggregatedTVL.Timestamp,
		"protocols": h.formatProtocolsTVL(aggregatedTVL.Protocols),
		"chains":    h.formatChainTotals(aggregatedTVL.ChainTotals),
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache for 1 minute
	if h.cache != nil {
		h.cache.Set(cacheKey, jsonData, 1*time.Minute)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(jsonData)
}

// GET /api/v1/tvl/{protocol}
func (h *Handler) GetProtocolTVL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	protocol := vars["protocol"]
	ctx := r.Context()

	// Check cache
	cacheKey := fmt.Sprintf("tvl:protocol:%s", protocol)
	if h.cache != nil {
		if cached, err := h.cache.Get(cacheKey); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.Write(cached)
			return
		}
	}

	// Calculate protocol TVL
	tvlData, err := h.calculator.CalculateTVL(ctx, protocol)
	if err != nil {
		h.sendError(w, fmt.Sprintf("Protocol not found: %s", protocol), http.StatusNotFound)
		return
	}

	// Convert to JSON-friendly format
	totalFloat, _ := tvlData.TotalUSD.Float64()

	response := map[string]interface{}{
		"protocol":  tvlData.Protocol,
		"total_tvl": totalFloat,
		"timestamp": tvlData.Timestamp,
		"chains":    h.formatChainsTVL(tvlData.Chains),
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Cache for 1 minute
	if h.cache != nil {
		h.cache.Set(cacheKey, jsonData, 1*time.Minute)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(jsonData)
}

// GET /api/v1/tvl/{protocol}/history
func (h *Handler) GetHistoricalTVL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	protocol := vars["protocol"]
	ctx := r.Context()

	// Parse query parameters
	chain := r.URL.Query().Get("chain")
	period := r.URL.Query().Get("period") // 1h, 24h, 7d, 30d

	// Calculate time range based on period
	to := time.Now()
	from := to.Add(-24 * time.Hour) // Default 24h

	switch period {
	case "1h":
		from = to.Add(-1 * time.Hour)
	case "24h":
		from = to.Add(-24 * time.Hour)
	case "7d":
		from = to.Add(-7 * 24 * time.Hour)
	case "30d":
		from = to.Add(-30 * 24 * time.Hour)
	}

	// Get historical data
	snapshots, err := h.storage.GetHistoricalTVL(ctx, protocol, chain, from, to)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format response
	history := make([]map[string]interface{}, 0, len(snapshots))
	for _, snapshot := range snapshots {
		tvlFloat, _ := snapshot.TotalUSD.Float64()
		history = append(history, map[string]interface{}{
			"timestamp": snapshot.Timestamp,
			"tvl":       tvlFloat,
		})
	}

	response := map[string]interface{}{
		"protocol": protocol,
		"chain":    chain,
		"period":   period,
		"history":  history,
	}

	h.sendJSON(w, response)
}

// GET /api/v1/protocols
func (h *Handler) GetProtocols(w http.ResponseWriter, r *http.Request) {
	// This would normally fetch from storage
	// For now, return registered protocols
	protocols := []map[string]interface{}{
		{
			"name":        "uniswap-v2",
			"type":        "dex",
			"description": "Uniswap V2 DEX",
			"chains":      []string{"ethereum"},
		},
		// Add more protocols here
	}

	response := map[string]interface{}{
		"protocols": protocols,
		"count":     len(protocols),
	}

	h.sendJSON(w, response)
}

// GET /api/v1/chains
func (h *Handler) GetChains(w http.ResponseWriter, r *http.Request) {
	chains := h.calculator.GetSupportedChains()

	response := map[string]interface{}{
		"chains": chains,
		"count":  len(chains),
	}

	h.sendJSON(w, response)
}

// GET /api/v1/health
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Check if calculator is working
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	h.sendJSON(w, health)
}

// GET /api/v1/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get aggregated TVL for stats
	aggregatedTVL, err := h.calculator.CalculateAllProtocolsTVL(ctx)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	totalFloat, _ := aggregatedTVL.TotalUSD.Float64()

	stats := map[string]interface{}{
		"total_tvl":      totalFloat,
		"protocol_count": len(aggregatedTVL.Protocols),
		"chain_count":    len(aggregatedTVL.ChainTotals),
		"last_updated":   aggregatedTVL.Timestamp,
	}

	h.sendJSON(w, stats)
}

// Helper functions
func (h *Handler) sendJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) sendError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func (h *Handler) formatChainsTVL(chains map[string]*models.ChainTVL) map[string]interface{} {
	result := make(map[string]interface{})

	for chain, tvl := range chains {
		totalFloat, _ := tvl.TotalUSD.Float64()

		assets := make([]map[string]interface{}, 0, len(tvl.Assets))
		for _, asset := range tvl.Assets {
			valueFloat, _ := asset.ValueUSD.Float64()
			assets = append(assets, map[string]interface{}{
				"token":     asset.Token.Hex(),
				"symbol":    asset.Symbol,
				"amount":    asset.Amount.String(),
				"decimals":  asset.Decimals,
				"value_usd": valueFloat,
			})
		}

		result[chain] = map[string]interface{}{
			"total_usd": totalFloat,
			"assets":    assets,
		}
	}

	return result
}

func (h *Handler) formatProtocolsTVL(protocols map[string]*models.TVLData) map[string]interface{} {
	result := make(map[string]interface{})

	for name, tvl := range protocols {
		totalFloat, _ := tvl.TotalUSD.Float64()
		result[name] = map[string]interface{}{
			"total_usd": totalFloat,
			"timestamp": tvl.Timestamp,
		}
	}

	return result
}

func (h *Handler) formatChainTotals(chains map[string]*big.Float) map[string]float64 {
	result := make(map[string]float64)

	for chain, total := range chains {
		totalFloat, _ := total.Float64()
		result[chain] = totalFloat
	}

	return result
}
