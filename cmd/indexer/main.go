// cmd/indexer/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
	"github.com/zacksfF/evm-tvl-aggregator/internal/indexer"
)

func main() {
	// Create blockchain manager
	manager := blockchain.NewManager()

	// Get RPC URL from environment or use Alchemy default
	rpcURL := os.Getenv("ETH_RPC_URL")
	if rpcURL == "" {
		// Public nodes have restrictions on eth_getLogs, use Alchemy for indexing
		rpcURL = "https://eth-mainnet.g.alchemy.com/v2/key"
		log.Println("Using Alchemy RPC for indexing (required for eth_getLogs without address filter)")
	}

	// Add Ethereum mainnet
	err := manager.AddChain(blockchain.ChainConfig{
		Name:    "ethereum",
		ChainID: big.NewInt(1),
		RPCURL:  rpcURL,
		WSURL:   os.Getenv("ETH_WS_URL"), // Optional: add WSS URL for real-time events
	})
	if err != nil {
		log.Fatal(err)
	}

	// Setup storage - try PostgreSQL first, fallback to memory
	var storage indexer.Storage

	// Check if we have database environment variables
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost != "" && dbUser != "" && dbName != "" {
		dbURL := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
			dbUser, dbPassword, dbHost, dbName)

		storage, err = indexer.NewPostgresStorage(dbURL)
		if err != nil {
			log.Printf("Failed to connect to PostgreSQL (%v), using in-memory storage", err)
			storage = indexer.NewMemoryStorage()
		} else {
			log.Println("Connected to PostgreSQL database")
		}
	} else {
		log.Println("No database configuration found, using in-memory storage")
		storage = indexer.NewMemoryStorage()
	}

	// Create indexer
	config := indexer.Config{
		BatchSize:          10, // Small batches for reliable processing
		WorkerCount:        1,  // Single worker to avoid rate limits
		BlockConfirmations: 12,
		RetryAttempts:      3,
		RetryDelay:         5 * time.Second,
		StartFromBlock:     100, // Index last 100 blocks
	}

	idx := indexer.NewIndexer(manager, storage, config)

	// Register Uniswap V2 processor
	uniswapFactory := common.HexToAddress("0x5C69bEe701ef814a2B6a3EDD4B1652CB9cc5aA6f")
	uniswapProcessor := indexer.NewUniswapV2Processor(uniswapFactory)
	idx.RegisterProcessor(uniswapProcessor)

	// Register Aave V3 processor
	aavePool := common.HexToAddress("0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2")
	aaveProcessor := indexer.NewAaveV3Processor(aavePool)
	idx.RegisterProcessor(aaveProcessor)

	// Start indexing
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := idx.Start(ctx); err != nil {
			log.Printf("Indexer error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nShutting down...")

	cancel()
	idx.Stop()

	fmt.Println("Indexer stopped successfully")
}
