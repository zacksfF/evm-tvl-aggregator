package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/evm-tvl-aggregator/internal/aggregator"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage/memory"
)

func main() {
	// Create blockchain manager
	manager := blockchain.NewManager()

	// Add Ethereum mainnet
	err := manager.AddChain(blockchain.ChainConfig{
		Name:        "ethereum",
		ChainID:     big.NewInt(1),
		RPCURL:      "https://eth-mainnet.g.alchemy.com/v2/key",
		NativeToken: "ETH",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Create price oracle with mock prices for testing
	priceOracle := aggregator.NewPriceOracle()
	priceOracle.SetMockPrice("ETH", 2500)
	priceOracle.SetMockPrice("USDC", 1)
	priceOracle.SetMockPrice("USDT", 1)
	priceOracle.SetMockPrice("DAI", 1)
	priceOracle.SetMockPrice("WETH", 2500)

	// Create storage
	storage := memory.NewMemoryStorage()

	// Create TVL calculator
	calculator := aggregator.NewTVLCalculator(manager, priceOracle, storage)

	// Register a test protocol (Uniswap V2)
	uniswapProtocol := &models.Protocol{
		Name:        "uniswap-v2",
		Type:        models.ProtocolTypeDEX,
		Description: "Uniswap V2 DEX",
		Website:     "https://uniswap.org",
		Chains: map[string][]models.ContractConfig{
			"ethereum": {
				{
					Address: common.HexToAddress("0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"), // USDC/WETH pool
					Name:    "USDC-WETH",
					Type:    "pool",
				},
			},
		},
	}

	calculator.RegisterProtocol(uniswapProtocol)

	// Calculate TVL
	ctx := context.Background()

	fmt.Println("Calculating TVL for Uniswap V2...")
	tvl, err := calculator.CalculateTVL(ctx, "uniswap-v2")
	if err != nil {
		log.Printf("Error calculating TVL: %v", err)
	} else {
		total, _ := tvl.TotalUSD.Float64()
		fmt.Printf("\nTVL Results:\n")
		fmt.Printf("Protocol: %s\n", tvl.Protocol)
		fmt.Printf("Total TVL: $%.2f\n", total)

		for chain, chainTVL := range tvl.Chains {
			chainTotal, _ := chainTVL.TotalUSD.Float64()
			fmt.Printf("\n%s Chain TVL: $%.2f\n", chain, chainTotal)

			for _, asset := range chainTVL.Assets {
				assetValue, _ := asset.ValueUSD.Float64()
				fmt.Printf("  - %s: %.4f tokens ($%.2f)\n",
					asset.Symbol,
					new(big.Float).SetInt(asset.Amount),
					assetValue,
				)
			}
		}
	}

	// Test aggregated TVL across all protocols
	fmt.Println("\n\nCalculating aggregated TVL...")
	aggregatedTVL, err := calculator.CalculateAllProtocolsTVL(ctx)
	if err != nil {
		log.Printf("Error calculating aggregated TVL: %v", err)
	} else {
		total, _ := aggregatedTVL.TotalUSD.Float64()
		fmt.Printf("\nTotal TVL across all protocols: $%.2f\n", total)

		for protocol, tvlData := range aggregatedTVL.Protocols {
			protocolTotal, _ := tvlData.TotalUSD.Float64()
			fmt.Printf("  %s: $%.2f\n", protocol, protocolTotal)
		}
	}

	fmt.Println("\nDone!")
}
