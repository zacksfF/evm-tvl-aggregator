package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// Contract represents a smart contract
type Contract struct {
	Address common.Address
	ABI     abi.ABI
	Client  *Client
}

// NewContract creates a new contract instance
func NewContract(address common.Address, abiJSON string, client *Client) (*Contract, error) {
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return &Contract{
		Address: address,
		ABI:     parsedABI,
		Client:  client,
	}, nil
}

// Call executes a read-only contract call
func (c *Contract) Call(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	data, err := c.ABI.Pack(method, args...)
	if err != nil {
		return fmt.Errorf("failed to pack method %s: %w", method, err)
	}

	msg := ethereum.CallMsg{
		To:   &c.Address,
		Data: data,
	}

	output, err := c.Client.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return fmt.Errorf("failed to call contract: %w", err)
	}

	err = c.ABI.UnpackIntoInterface(result, method, output)
	if err != nil {
		return fmt.Errorf("failed to unpack result: %w", err)
	}

	return nil
}

// GetBalance gets the balance of a token for an address
func (c *Contract) GetBalance(ctx context.Context, tokenAddress, holderAddress common.Address) (*big.Int, error) {
	// ERC20 balanceOf method
	const balanceOfABI = `[{"constant":true,"inputs":[{"name":"","type":"address"}],"name":"balanceOf","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`

	erc20ABI, err := abi.JSON(strings.NewReader(balanceOfABI))
	if err != nil {
		return nil, err
	}

	data, err := erc20ABI.Pack("balanceOf", holderAddress)
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	output, err := c.Client.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	var balance *big.Int
	err = erc20ABI.UnpackIntoInterface(&balance, "balanceOf", output)
	if err != nil {
		return nil, err
	}

	return balance, nil
}

// GetTotalSupply gets the total supply of a token
func (c *Contract) GetTotalSupply(ctx context.Context, tokenAddress common.Address) (*big.Int, error) {
	const totalSupplyABI = `[{"constant":true,"inputs":[],"name":"totalSupply","outputs":[{"name":"","type":"uint256"}],"type":"function"}]`

	erc20ABI, err := abi.JSON(strings.NewReader(totalSupplyABI))
	if err != nil {
		return nil, err
	}

	data, err := erc20ABI.Pack("totalSupply")
	if err != nil {
		return nil, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	output, err := c.Client.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}

	var supply *big.Int
	err = erc20ABI.UnpackIntoInterface(&supply, "totalSupply", output)
	if err != nil {
		return nil, err
	}

	return supply, nil
}

// GetDecimals gets the decimals of a token
func (c *Contract) GetDecimals(ctx context.Context, tokenAddress common.Address) (uint8, error) {
	const decimalsABI = `[{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"type":"function"}]`

	erc20ABI, err := abi.JSON(strings.NewReader(decimalsABI))
	if err != nil {
		return 0, err
	}

	data, err := erc20ABI.Pack("decimals")
	if err != nil {
		return 0, err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	output, err := c.Client.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return 0, err
	}

	var decimals uint8
	err = erc20ABI.UnpackIntoInterface(&decimals, "decimals", output)
	if err != nil {
		return 0, err
	}

	return decimals, nil
}

// GetSymbol gets the symbol of a token
func (c *Contract) GetSymbol(ctx context.Context, tokenAddress common.Address) (string, error) {
	const symbolABI = `[{"constant":true,"inputs":[],"name":"symbol","outputs":[{"name":"","type":"string"}],"type":"function"}]`

	erc20ABI, err := abi.JSON(strings.NewReader(symbolABI))
	if err != nil {
		return "", err
	}

	data, err := erc20ABI.Pack("symbol")
	if err != nil {
		return "", err
	}

	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	output, err := c.Client.GetClient().CallContract(ctx, msg, nil)
	if err != nil {
		return "", err
	}

	var symbol string
	err = erc20ABI.UnpackIntoInterface(&symbol, "symbol", output)
	if err != nil {
		return "", err
	}

	return symbol, nil
}
