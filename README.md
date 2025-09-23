# EVM TVL Aggregator

**Status: Under Development**

A high-performance, multi-chain Total Value Locked (TVL) aggregator for EVM-compatible blockchains. Built with Go for maximum efficiency and real-time data processing.

![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-development-orange)

## Overview

EVM TVL Aggregator is a comprehensive solution for tracking and calculating Total Value Locked across multiple DeFi protocols on EVM-compatible chains. It provides real-time TVL calculations, historical data tracking, and both REST API and Terminal UI interfaces.

### Key Features

- Multi-Chain Support: Ethereum, Polygon, Arbitrum, BSC, and more
- Real-Time TVL Calculation: Live tracking of protocol TVL with configurable updates
- Interactive Terminal UI: Beautiful command-line interface with real-time dashboards
- Protocol Adapters: Built-in support for Uniswap, Aave, Compound, and custom protocols
- Event Indexing: Efficient blockchain event scanning and storage
- Price Oracle Integration: Multiple price sources (CoinGecko, Chainlink, DEX prices)
- High Performance: Concurrent processing with Go routines
- RESTful API: Easy-to-use endpoints with response caching
- Flexible Storage: PostgreSQL for production, in-memory for testing
- Historical Data: Track TVL changes over time
- Configurable: YAML configuration with CLI overrides

## Demo Video

[![Terminal Demo](https://img.shields.io/badge/Demo-Terminal%20UI-blue)](https://github.com/yourusername/evm-tvl-aggregator)

*A screen recording of the terminal interface will be added here showing the real-time TVL monitoring, interactive navigation, and data visualization features.*

## RPC Node Requirements

**Important**: This application requires specific RPC node capabilities depending on usage:

### For API Server (Basic Queries)
- **Public RPC nodes**: Compatible with most public endpoints
- **Rate limits**: Can handle standard rate-limited public RPCs
- **Features needed**: Basic `eth_call`, `eth_getBlockByNumber`

### For Indexer (Event Scanning)
- **Full-featured RPC required**: Alchemy, Infura, QuickNode, or self-hosted
- **eth_getLogs support**: Must support unrestricted log filtering
- **Rate limits**: Higher limits needed for batch processing
- **Websocket support**: Recommended for real-time event streaming

### Recommended RPC Providers
- **Alchemy**: Best performance, requires API key
- **Infura**: Reliable, requires API key  
- **QuickNode**: Fast, requires subscription
- **Public nodes**: Limited functionality, good for testing

## Quick Start

### Prerequisites
- Go 1.21 or higher
- PostgreSQL 14+ (optional, for production)
- Ethereum RPC endpoint

### Installation

```bash
# Clone repository
git clone https://github.com/zacksfF/evm-tvl-aggregator.git
cd evm-tvl-aggregator

# Install dependencies
go mod download

# Copy environment configuration
cp .env.example .env
# Edit .env with your RPC endpoints
```

### Running

#### Terminal Interface
```bash
# Start API server
go run cmd/api/main.go

# In another terminal, start TUI
go run cmd/tui/main.go

# Or use the demo script
./run_terminal.sh
```

#### API Server Only
```bash
go run cmd/api/main.go
# API available at http://localhost:8080
```

#### With Docker
```bash
docker-compose up
```

## API Documentation

### Base URL
```
http://localhost:8080/api/v1
```

### Core Endpoints

**Health Check**
```http
GET /health
```

**Total TVL**
```http
GET /tvl
```
Returns aggregated TVL across all protocols and chains.

**Protocol TVL**
```http
GET /tvl/{protocol}
```
Returns TVL for a specific protocol (e.g., `uniswap-v2`).

**Supported Protocols**
```http
GET /protocols
```

**Supported Chains**
```http
GET /chains
```

**System Statistics**
```http
GET /stats
```

### Response Format
```json
{
  "total_tvl": 28600378.57,
  "timestamp": "2025-09-23T10:00:00Z",
  "protocols": {
    "uniswap-v2": {
      "total_usd": 28600378.57,
      "timestamp": "2025-09-23T10:00:00Z"
    }
  },
  "chains": {
    "ethereum": 28600378.57
  }
}
```

## Configuration

### Environment Variables

Create a `.env` file:

```env
# Database (optional)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=tvl_aggregator

# Ethereum RPC
ETH_RPC_URL=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
ETH_WS_URL=wss://eth-mainnet.g.alchemy.com/v2/YOUR_KEY

# API Configuration
API_PORT=8080
API_RATE_LIMIT=100

# Indexer Configuration
INDEXER_BATCH_SIZE=100
INDEXER_WORKERS=3
```

### TUI Configuration

The terminal interface can be configured via `.tvl-aggregator.yaml`:

```yaml
api:
  url: "http://localhost:8080"
  timeout: "30s"

ui:
  refresh_interval: "5s"
  theme: "dark"
  no_color: false

log:
  level: "info"
```

## Adding New Protocols

To add support for a new DeFi protocol:

1. **Define Protocol Configuration**
```go
protocol := &aggregator.Protocol{
    Name:        "aave-v3",
    Type:        "lending",
    Description: "Aave V3 Lending Protocol",
    Website:     "https://aave.com",
    Chains: map[string][]aggregator.ContractConfig{
        "ethereum": {
            {
                Address: "0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2",
                Name:    "Aave V3 Pool",
                Type:    "lending-pool",
            },
        },
    },
}
```

2. **Register Protocol**
```go
calculator.RegisterProtocol(protocol)
```

## Adding New Chains

To add support for a new blockchain:

```go
err := manager.AddChain(blockchain.ChainConfig{
    Name:        "polygon",
    ChainID:     big.NewInt(137),
    RPCURL:      os.Getenv("POLYGON_RPC_URL"),
    NativeToken: "MATIC",
})
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Built with Go for the DeFi community
