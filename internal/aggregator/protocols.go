// internal/aggregator/protocols.go
package aggregator

import (
	"context"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/zacksfF/evm-tvl-aggregator/internal/blockchain"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
)

// Uniswap V2 TVL Calculator
func (tc *TVLCalculator) calculateUniswapV2TVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// Uniswap V2 Pair ABI for getReserves
	const pairABI = `[{"constant":true,"inputs":[],"name":"getReserves","outputs":[{"name":"reserve0","type":"uint112"},{"name":"reserve1","type":"uint112"},{"name":"blockTimestampLast","type":"uint32"}],"type":"function"},{"constant":true,"inputs":[],"name":"token0","outputs":[{"name":"","type":"address"}],"type":"function"},{"constant":true,"inputs":[],"name":"token1","outputs":[{"name":"","type":"address"}],"type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(pairABI))
	if err != nil {
		return nil, err
	}

	pairAddr := contract.Address

	// Get token addresses
	token0Data, err := parsedABI.Pack("token0")
	if err != nil {
		return nil, err
	}

	token0Result, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &pairAddr,
		Data: token0Data,
	}, nil)
	if err != nil {
		return nil, err
	}

	token1Data, err := parsedABI.Pack("token1")
	if err != nil {
		return nil, err
	}

	token1Result, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &pairAddr,
		Data: token1Data,
	}, nil)
	if err != nil {
		return nil, err
	}

	var token0, token1 common.Address
	parsedABI.UnpackIntoInterface(&token0, "token0", token0Result)
	parsedABI.UnpackIntoInterface(&token1, "token1", token1Result)

	// Get reserves
	reservesData, err := parsedABI.Pack("getReserves")
	if err != nil {
		return nil, err
	}

	result, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &pairAddr,
		Data: reservesData,
	}, nil)
	if err != nil {
		return nil, err
	}

	var reserves struct {
		Reserve0           *big.Int
		Reserve1           *big.Int
		BlockTimestampLast uint32
	}

	err = parsedABI.UnpackIntoInterface(&reserves, "getReserves", result)
	if err != nil {
		return nil, err
	}

	// Get token info and prices
	basicContract := &blockchain.Contract{Client: client}

	symbol0, _ := basicContract.GetSymbol(ctx, token0)
	symbol1, _ := basicContract.GetSymbol(ctx, token1)
	decimals0, _ := basicContract.GetDecimals(ctx, token0)
	decimals1, _ := basicContract.GetDecimals(ctx, token1)

	if decimals0 == 0 {
		decimals0 = 18
	}
	if decimals1 == 0 {
		decimals1 = 18
	}

	price0, _ := tc.priceOracle.GetTokenPrice(ctx, token0.Hex())
	price1, _ := tc.priceOracle.GetTokenPrice(ctx, token1.Hex())

	// Calculate USD values
	divisor0 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals0)), nil)
	divisor1 := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals1)), nil)

	value0 := new(big.Float).SetInt(reserves.Reserve0)
	value0.Mul(value0, price0)
	value0.Quo(value0, new(big.Float).SetInt(divisor0))

	value1 := new(big.Float).SetInt(reserves.Reserve1)
	value1.Mul(value1, price1)
	value1.Quo(value1, new(big.Float).SetInt(divisor1))

	return []*models.AssetTVL{
		{
			Token:    token0,
			Symbol:   symbol0,
			Amount:   reserves.Reserve0,
			Decimals: decimals0,
			ValueUSD: value0,
		},
		{
			Token:    token1,
			Symbol:   symbol1,
			Amount:   reserves.Reserve1,
			Decimals: decimals1,
			ValueUSD: value1,
		},
	}, nil
}

// Aave V3 TVL Calculator
func (tc *TVLCalculator) calculateAaveV3TVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// Aave V3 Pool ABI for getReservesList and getReserveData
	const poolABI = `[{"inputs":[],"name":"getReservesList","outputs":[{"internalType":"address[]","name":"","type":"address[]"}],"stateMutability":"view","type":"function"},{"inputs":[{"internalType":"address","name":"asset","type":"address"}],"name":"getReserveData","outputs":[{"components":[{"internalType":"uint256","name":"configuration","type":"uint256"},{"internalType":"uint128","name":"liquidityIndex","type":"uint128"},{"internalType":"uint128","name":"currentLiquidityRate","type":"uint128"},{"internalType":"uint128","name":"variableBorrowIndex","type":"uint128"},{"internalType":"uint128","name":"currentVariableBorrowRate","type":"uint128"},{"internalType":"uint128","name":"currentStableBorrowRate","type":"uint128"},{"internalType":"uint40","name":"lastUpdateTimestamp","type":"uint40"},{"internalType":"uint16","name":"id","type":"uint16"},{"internalType":"address","name":"aTokenAddress","type":"address"},{"internalType":"address","name":"stableDebtTokenAddress","type":"address"},{"internalType":"address","name":"variableDebtTokenAddress","type":"address"},{"internalType":"address","name":"interestRateStrategyAddress","type":"address"},{"internalType":"uint128","name":"accruedToTreasury","type":"uint128"},{"internalType":"uint128","name":"unbacked","type":"uint128"},{"internalType":"uint128","name":"isolationModeTotalDebt","type":"uint128"}],"internalType":"struct DataTypes.ReserveData","name":"","type":"tuple"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(poolABI))
	if err != nil {
		return nil, err
	}

	poolAddr := contract.Address

	// Get reserves list
	reservesData, err := parsedABI.Pack("getReservesList")
	if err != nil {
		return nil, err
	}

	result, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &poolAddr,
		Data: reservesData,
	}, nil)
	if err != nil {
		return nil, err
	}

	var reserves []common.Address
	err = parsedABI.UnpackIntoInterface(&reserves, "getReservesList", result)
	if err != nil {
		return nil, err
	}

	assets := make([]*models.AssetTVL, 0)
	basicContract := &blockchain.Contract{Client: client}

	// Get TVL for each reserve
	for _, reserve := range reserves {
		// Get aToken address from reserve data
		reserveData, err := parsedABI.Pack("getReserveData", reserve)
		if err != nil {
			continue
		}

		reserveResult, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
			To:   &poolAddr,
			Data: reserveData,
		}, nil)
		if err != nil {
			continue
		}

		// Parse to get aToken address (9th field in the struct)
		if len(reserveResult) < 288 { // Minimum size for the struct
			continue
		}

		aTokenAddr := common.BytesToAddress(reserveResult[256:288])

		// Get total supply of aToken
		totalSupply, err := basicContract.GetTotalSupply(ctx, aTokenAddr)
		if err != nil || totalSupply.Cmp(big.NewInt(0)) == 0 {
			continue
		}

		// Get token metadata
		symbol, _ := basicContract.GetSymbol(ctx, reserve)
		decimals, _ := basicContract.GetDecimals(ctx, reserve)
		if decimals == 0 {
			decimals = 18
		}

		// Get price
		price, err := tc.priceOracle.GetTokenPrice(ctx, reserve.Hex())
		if err != nil {
			price = big.NewFloat(0)
		}

		// Calculate USD value
		divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
		valueUSD := new(big.Float).SetInt(totalSupply)
		valueUSD.Mul(valueUSD, price)
		valueUSD.Quo(valueUSD, new(big.Float).SetInt(divisor))

		assets = append(assets, &models.AssetTVL{
			Token:    reserve,
			Symbol:   symbol,
			Amount:   totalSupply,
			Decimals: decimals,
			ValueUSD: valueUSD,
		})
	}

	return assets, nil
}

// Compound V3 TVL Calculator
func (tc *TVLCalculator) calculateCompoundV3TVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// Simplified Compound V3 calculation
	// Get total supply and borrows for the base asset

	const cometABI = `[{"inputs":[],"name":"totalSupply","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"totalBorrow","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"baseToken","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(cometABI))
	if err != nil {
		return nil, err
	}

	cometAddr := contract.Address

	// Get base token
	baseTokenData, err := parsedABI.Pack("baseToken")
	if err != nil {
		return nil, err
	}

	baseTokenResult, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &cometAddr,
		Data: baseTokenData,
	}, nil)
	if err != nil {
		return nil, err
	}

	var baseToken common.Address
	parsedABI.UnpackIntoInterface(&baseToken, "baseToken", baseTokenResult)

	// Get total supply
	supplyData, err := parsedABI.Pack("totalSupply")
	if err != nil {
		return nil, err
	}

	supplyResult, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &cometAddr,
		Data: supplyData,
	}, nil)
	if err != nil {
		return nil, err
	}

	totalSupply := new(big.Int).SetBytes(supplyResult)

	// Get token info
	basicContract := &blockchain.Contract{Client: client}
	symbol, _ := basicContract.GetSymbol(ctx, baseToken)
	decimals, _ := basicContract.GetDecimals(ctx, baseToken)
	if decimals == 0 {
		decimals = 18
	}

	// Get price
	price, err := tc.priceOracle.GetTokenPrice(ctx, baseToken.Hex())
	if err != nil {
		price = big.NewFloat(0)
	}

	// Calculate USD value
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	valueUSD := new(big.Float).SetInt(totalSupply)
	valueUSD.Mul(valueUSD, price)
	valueUSD.Quo(valueUSD, new(big.Float).SetInt(divisor))

	return []*models.AssetTVL{
		{
			Token:    baseToken,
			Symbol:   symbol,
			Amount:   totalSupply,
			Decimals: decimals,
			ValueUSD: valueUSD,
		},
	}, nil
}

// DEX TVL Calculator (generic for Uniswap-like DEXs)
func (tc *TVLCalculator) calculateDexTVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// Try Uniswap V2 style first
	if assets, err := tc.calculateUniswapV2TVL(ctx, client, contract); err == nil {
		return assets, nil
	}

	// Fallback to generic TVL calculation
	return tc.calculateGenericTVL(ctx, client, contract)
}

// Lending TVL Calculator
func (tc *TVLCalculator) calculateLendingTVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// Try Aave V3 first
	if strings.Contains(contract.Name, "aave") {
		if assets, err := tc.calculateAaveV3TVL(ctx, client, contract); err == nil {
			return assets, nil
		}
	}

	// Try Compound V3
	if strings.Contains(contract.Name, "compound") {
		if assets, err := tc.calculateCompoundV3TVL(ctx, client, contract); err == nil {
			return assets, nil
		}
	}

	// Fallback to generic
	return tc.calculateGenericTVL(ctx, client, contract)
}

// Yield TVL Calculator
func (tc *TVLCalculator) calculateYieldTVL(ctx context.Context, client *blockchain.Client, contract models.ContractConfig) ([]*models.AssetTVL, error) {
	// For yield protocols, typically check vault balances
	// This is a simplified version

	const vaultABI = `[{"inputs":[],"name":"totalAssets","outputs":[{"internalType":"uint256","name":"","type":"uint256"}],"stateMutability":"view","type":"function"},{"inputs":[],"name":"asset","outputs":[{"internalType":"address","name":"","type":"address"}],"stateMutability":"view","type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(vaultABI))
	if err != nil {
		// Fallback to generic if not a standard vault
		return tc.calculateGenericTVL(ctx, client, contract)
	}

	vaultAddr := contract.Address

	// Try to get total assets
	assetsData, err := parsedABI.Pack("totalAssets")
	if err != nil {
		return tc.calculateGenericTVL(ctx, client, contract)
	}

	result, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &vaultAddr,
		Data: assetsData,
	}, nil)
	if err != nil {
		return tc.calculateGenericTVL(ctx, client, contract)
	}

	totalAssets := new(big.Int).SetBytes(result)

	// Get underlying asset
	assetData, err := parsedABI.Pack("asset")
	if err != nil {
		return tc.calculateGenericTVL(ctx, client, contract)
	}

	assetResult, err := client.GetClient().CallContract(ctx, ethereum.CallMsg{
		To:   &vaultAddr,
		Data: assetData,
	}, nil)
	if err != nil {
		return tc.calculateGenericTVL(ctx, client, contract)
	}

	var asset common.Address
	parsedABI.UnpackIntoInterface(&asset, "asset", assetResult)

	// Get token info
	basicContract := &blockchain.Contract{Client: client}
	symbol, _ := basicContract.GetSymbol(ctx, asset)
	decimals, _ := basicContract.GetDecimals(ctx, asset)
	if decimals == 0 {
		decimals = 18
	}

	// Get price
	price, err := tc.priceOracle.GetTokenPrice(ctx, asset.Hex())
	if err != nil {
		price = big.NewFloat(0)
	}

	// Calculate USD value
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	valueUSD := new(big.Float).SetInt(totalAssets)
	valueUSD.Mul(valueUSD, price)
	valueUSD.Quo(valueUSD, new(big.Float).SetInt(divisor))

	return []*models.AssetTVL{
		{
			Token:    asset,
			Symbol:   symbol,
			Amount:   totalAssets,
			Decimals: decimals,
			ValueUSD: valueUSD,
		},
	}, nil
}
