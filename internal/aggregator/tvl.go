package aggregator

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage"
)

type TVLCalculator struct {
	manager     *blockchain.Manager
	priceOracle *PriceOracle
	protocols   map[string]*models.Protocol
	storage     storage.Storage
	cache       Cache
	mu          sync.RWMutex
}


type Cache interface {
	Get(key string) (*models.TVLData, error)
	Set(key string, data *models.TVLData, ttl time.Duration) error
	Delete(key string) error
}

func NewTVLCalculator(manager *blockchain.Manager, priceOracle *PriceOracle, storage storage.Storage) *TVLCalculator {
	return &TVLCalculator{
		manager:     manager,
		priceOracle: priceOracle,
		protocols:   make(map[string]*models.Protocol),
		storage:     storage,
	}
}

func (tc *TVLCalculator) RegisterProtocol(protocol *models.Protocol) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.protocols[protocol.Name] = protocol
	fmt.Printf("Registered protocol: %s\n", protocol.Name)
}

func (tc *TVLCalculator) CalculateTVL(ctx context.Context, protocolName string) (*models.TVLData, error) {
	tc.mu.RLock()
	protocol, exists := tc.protocols[protocolName]
	tc.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("protocol %s not found", protocolName)
	}

	// Check cache first
	cacheKey := fmt.Sprintf("tvl:%s:%d", protocolName, time.Now().Unix()/60) // 1-minute cache
	if tc.cache != nil {
		if cached, err := tc.cache.Get(cacheKey); err == nil {
			return cached, nil
		}
	}

	tvlData := &models.TVLData{
		Protocol:  protocolName,
		Timestamp: time.Now(),
		Chains:    make(map[string]*models.ChainTVL),
		TotalUSD:  big.NewFloat(0),
	}

	// Calculate TVL for each chain
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	for chain, contracts := range protocol.Chains {
		wg.Add(1)
		go func(chainName string, contractAddrs []models.ContractConfig) {
			defer wg.Done()

			chainTVL, err := tc.calculateChainTVL(ctx, protocol, chainName, contractAddrs)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("chain %s: %w", chainName, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			tvlData.Chains[chainName] = chainTVL
			tvlData.TotalUSD.Add(tvlData.TotalUSD, chainTVL.TotalUSD)
			mu.Unlock()
		}(chain, contracts)
	}

	wg.Wait()

	if len(errors) > 0 {
		return nil, fmt.Errorf("failed to calculate TVL: %v", errors)
	}

	// Cache the result
	if tc.cache != nil {
		tc.cache.Set(cacheKey, tvlData, 1*time.Minute)
	}

	// Save snapshot
	snapshot := &models.TVLSnapshot{
		Protocol:  protocolName,
		Timestamp: tvlData.Timestamp,
		TotalUSD:  tvlData.TotalUSD,
	}

	if err := tc.storage.SaveTVLSnapshot(ctx, snapshot); err != nil {
		// Log but don't fail
		fmt.Printf("Failed to save TVL snapshot: %v\n", err)
	}

	return tvlData, nil
}

func (tc *TVLCalculator) calculateChainTVL(ctx context.Context, protocol *models.Protocol, chain string, contracts []models.ContractConfig) (*models.ChainTVL, error) {
	client, err := tc.manager.GetClient(chain)
	if err != nil {
		return nil, err
	}

	chainTVL := &models.ChainTVL{
		Chain:    chain,
		Assets:   make([]*models.AssetTVL, 0),
		TotalUSD: big.NewFloat(0),
	}

	for _, contract := range contracts {
		var assetTVLs []*models.AssetTVL

		switch protocol.Type {
		case "lending":
			assetTVLs, err = tc.calculateLendingTVL(ctx, client, contract)
		case "dex":
			assetTVLs, err = tc.calculateDexTVL(ctx, client, contract)
		case "yield":
			assetTVLs, err = tc.calculateYieldTVL(ctx, client, contract)
		default:
			assetTVLs, err = tc.calculateGenericTVL(ctx, client, contract)
		}

		if err != nil {
			fmt.Printf("Warning: Failed to calculate TVL for %s on %s: %v\n", contract.Address, chain, err)
			continue
		}

		for _, assetTVL := range assetTVLs {
			chainTVL.Assets = append(chainTVL.Assets, assetTVL)
			chainTVL.TotalUSD.Add(chainTVL.TotalUSD, assetTVL.ValueUSD)
		}
	}

	return chainTVL, nil
}

func (tc *TVLCalculator) calculateGenericTVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	assets := make([]*models.AssetTVL, 0)

	// Get native token balance
	balance, err := client.GetBalance(ctx, contract.Address)
	if err != nil {
		return nil, err
	}

	if balance.Cmp(big.NewInt(0)) > 0 {
		// Get native token price
		chainConfig, _ := tc.manager.GetChainConfig(client.GetChainName())
		price, err := tc.priceOracle.GetPrice(ctx, chainConfig.NativeToken)
		if err != nil {
			price = big.NewFloat(0) // Default to 0 if price not found
		}

		valueUSD := new(big.Float).SetInt(balance)
		valueUSD.Mul(valueUSD, price)
		valueUSD.Quo(valueUSD, big.NewFloat(1e18))

		assets = append(assets, &models.AssetTVL{
			Token:    common.HexToAddress("0x0"), // Native token
			Symbol:   chainConfig.NativeToken,
			Amount:   balance,
			Decimals: 18,
			ValueUSD: valueUSD,
		})
	}

	// Get tracked token balances
	if len(contract.Tokens) > 0 {
		tokenAssets, err := tc.getTokenBalances(ctx, client, contract)
		if err == nil {
			assets = append(assets, tokenAssets...)
		}
	}

	return assets, nil
}

func (tc *TVLCalculator) getTokenBalances(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	assets := make([]*models.AssetTVL, 0)
	contractAddr := contract.Address

	// Create a basic contract instance for token queries
	basicContract := &blockchain.Contract{
		Address: contractAddr,
		Client:  client,
	}

	for _, token := range contract.Tokens {

		// Get token balance
		balance, err := basicContract.GetBalance(ctx, token, contractAddr)
		if err != nil {
			continue
		}

		if balance.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		// Get token metadata
		symbol, _ := basicContract.GetSymbol(ctx, token)
		decimals, _ := basicContract.GetDecimals(ctx, token)
		if decimals == 0 {
			decimals = 18 // Default
		}

		// Get token price
		price, err := tc.priceOracle.GetTokenPrice(ctx, token.Hex())
		if err != nil {
			price = big.NewFloat(0)
		}

		// Calculate USD value
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		valueUSD := new(big.Float).SetInt(balance)
		valueUSD.Mul(valueUSD, price)
		valueUSD.Quo(valueUSD, new(big.Float).SetInt(divisor))

		assets = append(assets, &models.AssetTVL{
			Token:    token,
			Symbol:   symbol,
			Amount:   balance,
			Decimals: decimals,
			ValueUSD: valueUSD,
		})
	}

	return assets, nil
}

func (tc *TVLCalculator) CalculateAllProtocolsTVL(ctx context.Context) (*models.AggregatedTVL, error) {
	tc.mu.RLock()
	protocolNames := make([]string, 0, len(tc.protocols))
	for name := range tc.protocols {
		protocolNames = append(protocolNames, name)
	}
	tc.mu.RUnlock()

	aggregated := &models.AggregatedTVL{
		Timestamp:   time.Now(),
		Protocols:   make(map[string]*models.TVLData),
		TotalUSD:    big.NewFloat(0),
		ChainTotals: make(map[string]*big.Float),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, protocolName := range protocolNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()

			tvl, err := tc.CalculateTVL(ctx, name)
			if err != nil {
				fmt.Printf("Failed to calculate TVL for %s: %v\n", name, err)
				return
			}

			mu.Lock()
			aggregated.Protocols[name] = tvl
			aggregated.TotalUSD.Add(aggregated.TotalUSD, tvl.TotalUSD)

			// Update chain totals
			for chain, chainTVL := range tvl.Chains {
				if aggregated.ChainTotals[chain] == nil {
					aggregated.ChainTotals[chain] = big.NewFloat(0)
				}
				aggregated.ChainTotals[chain].Add(aggregated.ChainTotals[chain], chainTVL.TotalUSD)
			}
			mu.Unlock()
		}(protocolName)
	}

	wg.Wait()
	return aggregated, nil
}

func (tc *TVLCalculator) GetSupportedChains() []string {
	return tc.manager.GetSupportedChains()
}
