package aggregator

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"
)

type PriceOracle struct {
    cache      map[string]*PriceData
    mu         sync.RWMutex
    httpClient *http.Client
}

type PriceData struct {
    Symbol    string     `json:"symbol"`
    Price     *big.Float `json:"price"`
    Timestamp time.Time  `json:"timestamp"`
}

func NewPriceOracle() *PriceOracle {
    return &PriceOracle{
        cache: make(map[string]*PriceData),
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
        },
    }
}

func (po *PriceOracle) GetPrice(ctx context.Context, symbol string) (*big.Float, error) {
    symbol = strings.ToUpper(symbol)
    
    // Check cache (1 minute TTL)
    po.mu.RLock()
    if cached, exists := po.cache[symbol]; exists {
        if time.Since(cached.Timestamp) < time.Minute {
            po.mu.RUnlock()
            return new(big.Float).Set(cached.Price), nil
        }
    }
    po.mu.RUnlock()
    
    // Fetch from CoinGecko
    price, err := po.fetchFromCoinGecko(ctx, symbol)
    if err != nil {
        // Try backup source or return cached even if stale
        po.mu.RLock()
        if cached, exists := po.cache[symbol]; exists {
            po.mu.RUnlock()
            return new(big.Float).Set(cached.Price), nil
        }
        po.mu.RUnlock()
        return nil, err
    }
    
    // Update cache
    po.mu.Lock()
    po.cache[symbol] = &PriceData{
        Symbol:    symbol,
        Price:     price,
        Timestamp: time.Now(),
    }
    po.mu.Unlock()
    
    return price, nil
}

func (po *PriceOracle) GetTokenPrice(ctx context.Context, address string) (*big.Float, error) {
    // Normalize address
    address = strings.ToLower(address)
    
    // Check cache
    po.mu.RLock()
    if cached, exists := po.cache[address]; exists {
        if time.Since(cached.Timestamp) < time.Minute {
            po.mu.RUnlock()
            return new(big.Float).Set(cached.Price), nil
        }
    }
    po.mu.RUnlock()
    
    // Map common token addresses to symbols (simplified)
    tokenMap := map[string]string{
        "0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2": "ETH",  // WETH
        "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48": "USDC",
        "0xdac17f958d2ee523a2206206994597c13d831ec7": "USDT",
        "0x6b175474e89094c44da98b954eedeac495271d0f": "DAI",
        "0x2260fac5e5542a773aa44fbcfedf7c193bc2c599": "WBTC",
    }
    
    if symbol, exists := tokenMap[address]; exists {
        return po.GetPrice(ctx, symbol)
    }
    
    // For unknown tokens, return a default price or fetch from DEX
    return big.NewFloat(0), nil
}

func (po *PriceOracle) fetchFromCoinGecko(ctx context.Context, symbol string) (*big.Float, error) {
    // Map symbols to CoinGecko IDs
    idMap := map[string]string{
        "ETH":   "ethereum",
        "BTC":   "bitcoin",
        "WBTC":  "wrapped-bitcoin",
        "USDC":  "usd-coin",
        "USDT":  "tether",
        "DAI":   "dai",
        "MATIC": "matic-network",
        "ARB":   "arbitrum",
        "OP":    "optimism",
    }
    
    coinID, exists := idMap[symbol]
    if !exists {
        // Default to 1 USD for stablecoins
        if strings.Contains(symbol, "USD") {
            return big.NewFloat(1), nil
        }
        return big.NewFloat(0), fmt.Errorf("unknown token: %s", symbol)
    }
    
    // CoinGecko API (free tier)
    url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coinID)
    
    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := po.httpClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    var result map[string]map[string]float64
    if err := json.Unmarshal(body, &result); err != nil {
        return nil, err
    }
    
    if priceData, exists := result[coinID]; exists {
        if price, exists := priceData["usd"]; exists {
            return big.NewFloat(price), nil
        }
    }
    
    return nil, fmt.Errorf("price not found for %s", symbol)
}

func (po *PriceOracle) GetMultiplePrices(ctx context.Context, symbols []string) (map[string]*big.Float, error) {
    prices := make(map[string]*big.Float)
    
    var wg sync.WaitGroup
    var mu sync.Mutex
    
    for _, symbol := range symbols {
        wg.Add(1)
        go func(s string) {
            defer wg.Done()
            
            price, err := po.GetPrice(ctx, s)
            if err == nil {
                mu.Lock()
                prices[s] = price
                mu.Unlock()
            }
        }(symbol)
    }
    
    wg.Wait()
    return prices, nil
}

// Mock prices for testing
func (po *PriceOracle) SetMockPrice(symbol string, price float64) {
    po.mu.Lock()
    defer po.mu.Unlock()
    
    po.cache[strings.ToUpper(symbol)] = &PriceData{
        Symbol:    strings.ToUpper(symbol),
        Price:     big.NewFloat(price),
        Timestamp: time.Now(),
    }
}