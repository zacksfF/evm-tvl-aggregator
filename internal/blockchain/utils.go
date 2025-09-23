package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// WeiToEther converts Wei to Ether
func WeiToEther(wei *big.Int) *big.Float {
    ether := new(big.Float).SetInt(wei)
    ether.Quo(ether, big.NewFloat(1e18))
    return ether
}

// EtherToWei converts Ether to Wei
func EtherToWei(ether *big.Float) *big.Int {
    wei := new(big.Float).Mul(ether, big.NewFloat(1e18))
    result := new(big.Int)
    wei.Int(result)
    return result
}

// FormatUnits formats a big.Int with decimals
func FormatUnits(amount *big.Int, decimals uint8) *big.Float {
    divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
    result := new(big.Float).SetInt(amount)
    result.Quo(result, new(big.Float).SetInt(divisor))
    return result
}

// ParseUnits parses a value with decimals to big.Int
func ParseUnits(value string, decimals uint8) (*big.Int, error) {
    amount, ok := new(big.Float).SetString(value)
    if !ok {
        return nil, fmt.Errorf("invalid amount: %s", value)
    }
    
    multiplier := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
    amount.Mul(amount, new(big.Float).SetInt(multiplier))
    
    result := new(big.Int)
    amount.Int(result)
    return result, nil
}

// GetBlockByNumber fetches a block by number
func (c *Client) GetBlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
    return c.client.BlockByNumber(ctx, number)
}

// GetTransaction fetches a transaction by hash
func (c *Client) GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
    return c.client.TransactionByHash(ctx, txHash)
}

// GetTransactionReceipt fetches a transaction receipt
func (c *Client) GetTransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
    return c.client.TransactionReceipt(ctx, txHash)
}

// WaitForTransaction waits for a transaction to be mined
func (c *Client) WaitForTransaction(ctx context.Context, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()
    
    ticker := time.NewTicker(2 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("timeout waiting for transaction %s", txHash.Hex())
            
        case <-ticker.C:
            receipt, err := c.GetTransactionReceipt(ctx, txHash)
            if err == nil && receipt != nil {
                return receipt, nil
            }
        }
    }
}

// IsContractAddress checks if an address is a contract
func (c *Client) IsContractAddress(ctx context.Context, address common.Address) (bool, error) {
    code, err := c.client.CodeAt(ctx, address, nil)
    if err != nil {
        return false, err
    }
    return len(code) > 0, nil
}

// EstimateGas estimates gas for a transaction
func (c *Client) EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error) {
    return c.client.EstimateGas(ctx, msg)
}

// GetNonce gets the nonce for an address
func (c *Client) GetNonce(ctx context.Context, address common.Address) (uint64, error) {
    return c.client.PendingNonceAt(ctx, address)
}

// GetGasPrice gets the current gas price
func (c *Client) GetGasPrice(ctx context.Context) (*big.Int, error) {
    return c.client.SuggestGasPrice(ctx)
}

// ChainIDToName converts chain ID to common chain names
func ChainIDToName(chainID *big.Int) string {
    switch chainID.Int64() {
    case 1:
        return "ethereum"
    case 56:
        return "bsc"
    case 137:
        return "polygon"
    case 250:
        return "fantom"
    case 43114:
        return "avalanche"
    case 42161:
        return "arbitrum"
    case 10:
        return "optimism"
    case 8453:
        return "base"
    default:
        return fmt.Sprintf("chain-%d", chainID)
    }
}