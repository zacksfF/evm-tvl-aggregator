package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/evm-tvl-aggregator/internal/aggregator"
	"github.com/zacksfF/evm-tvl-aggregator/internal/api"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage/memory"
)

func main() {
	// Setup blockchain manager
	manager := blockchain.NewManager()

	// Get RPC URL from environment or use public default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://ethereum-rpc.publicnode.com"
	}

	// Add Ethereum mainnet
	err := manager.AddChain(blockchain.ChainConfig{
		Name:        "ethereum",
		ChainID:     big.NewInt(1),
		RPCURL:      rpcURL,
		NativeToken: "ETH",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Setup price oracle
	priceOracle := aggregator.NewPriceOracle()

	// Mock prices for testing
	priceOracle.SetMockPrice("ETH", 2500)
	priceOracle.SetMockPrice("USDC", 1)
	priceOracle.SetMockPrice("USDT", 1)
	priceOracle.SetMockPrice("WETH", 2500)

	// Setup storage
	storage := memory.NewMemoryStorage()

	// Create TVL calculator
	calculator := aggregator.NewTVLCalculator(manager, priceOracle, storage)

	// Register protocols
	uniswapProtocol := &models.Protocol{
		Name:        "uniswap-v2",
		Type:        models.ProtocolTypeDEX,
		Description: "Uniswap V2 DEX",
		Website:     "https://uniswap.org",
		Chains: map[string][]models.ContractConfig{
			"ethereum": {
				{
					Address: common.HexToAddress("0xB4e16d0168e52d35CaCD2c6185b44281Ec28C9Dc"), // USDC/WETH
					Name:    "USDC-WETH",
					Type:    "pool",
				},
			},
		},
	}

	calculator.RegisterProtocol(uniswapProtocol)

	// Create API handler
	handler := api.NewHandler(calculator, storage)

	// Create router
	router := api.NewRouter(handler)

	// Apply middleware
	router = api.LoggingMiddleware(router)
	router = api.TimeoutMiddleware(30 * time.Second)(router)

	// Get port from environment or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		fmt.Printf("🚀 TVL Aggregator API starting on http://localhost:%s\n", port)
		fmt.Println("\nEndpoints:")
		fmt.Println("  GET /api/v1/health           - Health check")
		fmt.Println("  GET /api/v1/tvl              - Total TVL across all protocols")
		fmt.Println("  GET /api/v1/tvl/{protocol}   - TVL for specific protocol")
		fmt.Println("  GET /api/v1/protocols        - List all protocols")
		fmt.Println("  GET /api/v1/chains           - List supported chains")
		fmt.Println("  GET /api/v1/stats            - System stats")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n⏸ Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	fmt.Println("✅ Server stopped")
}
