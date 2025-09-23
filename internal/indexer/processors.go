// internal/indexer/processors.go
package indexer

import (
    "fmt"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
)

// UniswapV2Processor processes Uniswap V2 events
type UniswapV2Processor struct {
    protocolName string
    factoryAddr  common.Address
}

func NewUniswapV2Processor(factory common.Address) *UniswapV2Processor {
    return &UniswapV2Processor{
        protocolName: "uniswap-v2",
        factoryAddr:  factory,
    }
}

func (p *UniswapV2Processor) GetProtocolName() string {
    return p.protocolName
}

func (p *UniswapV2Processor) GetEventSignatures() []common.Hash {
    return []common.Hash{
        crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)")),
        crypto.Keccak256Hash([]byte("Mint(address,uint256,uint256)")),
        crypto.Keccak256Hash([]byte("Burn(address,uint256,uint256,address)")),
        crypto.Keccak256Hash([]byte("Sync(uint112,uint112)")),
    }
}

func (p *UniswapV2Processor) Process(log types.Log) (*Event, error) {
    if len(log.Topics) == 0 {
        return nil, fmt.Errorf("no topics in log")
    }
    
    event := &Event{
        BlockNumber:     log.BlockNumber,
        BlockHash:       log.BlockHash.Hex(),
        TransactionHash: log.TxHash.Hex(),
        LogIndex:        log.Index,
        Address:         log.Address,
        Protocol:        p.protocolName,
    }
    
    switch log.Topics[0] {
    case crypto.Keccak256Hash([]byte("Swap(address,uint256,uint256,uint256,uint256,address)")):
        return p.processSwap(event, log)
    case crypto.Keccak256Hash([]byte("Mint(address,uint256,uint256)")):
        return p.processMint(event, log)
    case crypto.Keccak256Hash([]byte("Burn(address,uint256,uint256,address)")):
        return p.processBurn(event, log)
    case crypto.Keccak256Hash([]byte("Sync(uint112,uint112)")):
        return p.processSync(event, log)
    default:
        return nil, fmt.Errorf("unknown event signature")
    }
}

func (p *UniswapV2Processor) processSwap(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Swap"
    
    // Decode swap data
    if len(log.Data) < 128 {
        return nil, fmt.Errorf("invalid swap data")
    }
    
    amount0In := new(big.Int).SetBytes(log.Data[0:32])
    amount1In := new(big.Int).SetBytes(log.Data[32:64])
    amount0Out := new(big.Int).SetBytes(log.Data[64:96])
    amount1Out := new(big.Int).SetBytes(log.Data[96:128])
    
    event.Data = EventData{
        "sender":     common.BytesToAddress(log.Topics[1].Bytes()),
        "to":         common.BytesToAddress(log.Topics[2].Bytes()),
        "amount0_in":  amount0In.String(),
        "amount1_in":  amount1In.String(),
        "amount0_out": amount0Out.String(),
        "amount1_out": amount1Out.String(),
    }
    
    return event, nil
}

func (p *UniswapV2Processor) processMint(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Mint"
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("invalid mint data")
    }
    
    amount0 := new(big.Int).SetBytes(log.Data[0:32])
    amount1 := new(big.Int).SetBytes(log.Data[32:64])
    
    event.Data = EventData{
        "sender":  common.BytesToAddress(log.Topics[1].Bytes()),
        "amount0": amount0.String(),
        "amount1": amount1.String(),
    }
    
    return event, nil
}

func (p *UniswapV2Processor) processBurn(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Burn"
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("invalid burn data")
    }
    
    amount0 := new(big.Int).SetBytes(log.Data[0:32])
    amount1 := new(big.Int).SetBytes(log.Data[32:64])
    
    event.Data = EventData{
        "sender":  common.BytesToAddress(log.Topics[1].Bytes()),
        "to":      common.BytesToAddress(log.Topics[2].Bytes()),
        "amount0": amount0.String(),
        "amount1": amount1.String(),
    }
    
    return event, nil
}

func (p *UniswapV2Processor) processSync(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Sync"
    
    if len(log.Data) < 64 {
        return nil, fmt.Errorf("invalid sync data")
    }
    
    reserve0 := new(big.Int).SetBytes(log.Data[0:32])
    reserve1 := new(big.Int).SetBytes(log.Data[32:64])
    
    event.Data = EventData{
        "reserve0": reserve0.String(),
        "reserve1": reserve1.String(),
    }
    
    return event, nil
}

// AaveV3Processor processes Aave V3 events
type AaveV3Processor struct {
    protocolName string
    poolAddr     common.Address
}

func NewAaveV3Processor(pool common.Address) *AaveV3Processor {
    return &AaveV3Processor{
        protocolName: "aave-v3",
        poolAddr:     pool,
    }
}

func (p *AaveV3Processor) GetProtocolName() string {
    return p.protocolName
}

func (p *AaveV3Processor) GetEventSignatures() []common.Hash {
    return []common.Hash{
        crypto.Keccak256Hash([]byte("Supply(address,address,address,uint256,uint16)")),
        crypto.Keccak256Hash([]byte("Withdraw(address,address,address,uint256)")),
        crypto.Keccak256Hash([]byte("Borrow(address,address,address,uint256,uint8,uint256,uint16)")),
        crypto.Keccak256Hash([]byte("Repay(address,address,address,uint256,bool)")),
    }
}

func (p *AaveV3Processor) Process(log types.Log) (*Event, error) {
    if len(log.Topics) == 0 {
        return nil, fmt.Errorf("no topics in log")
    }
    
    event := &Event{
        BlockNumber:     log.BlockNumber,
        BlockHash:       log.BlockHash.Hex(),
        TransactionHash: log.TxHash.Hex(),
        LogIndex:        log.Index,
        Address:         log.Address,
        Protocol:        p.protocolName,
    }
    
    switch log.Topics[0] {
    case crypto.Keccak256Hash([]byte("Supply(address,address,address,uint256,uint16)")):
        return p.processSupply(event, log)
    case crypto.Keccak256Hash([]byte("Withdraw(address,address,address,uint256)")):
        return p.processWithdraw(event, log)
    case crypto.Keccak256Hash([]byte("Borrow(address,address,address,uint256,uint8,uint256,uint16)")):
        return p.processBorrow(event, log)
    case crypto.Keccak256Hash([]byte("Repay(address,address,address,uint256,bool)")):
        return p.processRepay(event, log)
    default:
        return nil, fmt.Errorf("unknown event signature")
    }
}

func (p *AaveV3Processor) processSupply(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Supply"
    
    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("invalid supply topics")
    }
    
    amount := new(big.Int).SetBytes(log.Data[0:32])
    
    event.Data = EventData{
        "reserve":   common.BytesToAddress(log.Topics[1].Bytes()),
        "user":      common.BytesToAddress(log.Topics[2].Bytes()),
        "on_behalf": common.BytesToAddress(log.Topics[3].Bytes()),
        "amount":    amount.String(),
    }
    
    return event, nil
}

func (p *AaveV3Processor) processWithdraw(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Withdraw"
    
    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("invalid withdraw topics")
    }
    
    amount := new(big.Int).SetBytes(log.Data[0:32])
    
    event.Data = EventData{
        "reserve": common.BytesToAddress(log.Topics[1].Bytes()),
        "user":    common.BytesToAddress(log.Topics[2].Bytes()),
        "to":      common.BytesToAddress(log.Topics[3].Bytes()),
        "amount":  amount.String(),
    }
    
    return event, nil
}

func (p *AaveV3Processor) processBorrow(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Borrow"
    
    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("invalid borrow topics")
    }
    
    amount := new(big.Int).SetBytes(log.Data[0:32])
    
    event.Data = EventData{
        "reserve":    common.BytesToAddress(log.Topics[1].Bytes()),
        "on_behalf":  common.BytesToAddress(log.Topics[2].Bytes()),
        "user":       common.BytesToAddress(log.Topics[3].Bytes()),
        "amount":     amount.String(),
    }
    
    return event, nil
}

func (p *AaveV3Processor) processRepay(event *Event, log types.Log) (*Event, error) {
    event.EventName = "Repay"
    
    if len(log.Topics) < 4 {
        return nil, fmt.Errorf("invalid repay topics")
    }
    
    amount := new(big.Int).SetBytes(log.Data[0:32])
    
    event.Data = EventData{
        "reserve": common.BytesToAddress(log.Topics[1].Bytes()),
        "user":    common.BytesToAddress(log.Topics[2].Bytes()),
        "repayer": common.BytesToAddress(log.Topics[3].Bytes()),
        "amount":  amount.String(),
    }
    
    return event, nil
}

// GenericERC20Processor for standard ERC20 transfers
type GenericERC20Processor struct {
    protocolName string
    tokens       []common.Address
}

func NewGenericERC20Processor(protocolName string, tokens []common.Address) *GenericERC20Processor {
    return &GenericERC20Processor{
        protocolName: protocolName,
        tokens:       tokens,
    }
}

func (p *GenericERC20Processor) GetProtocolName() string {
    return p.protocolName
}

func (p *GenericERC20Processor) GetEventSignatures() []common.Hash {
    return []common.Hash{
        crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)")),
    }
}

func (p *GenericERC20Processor) Process(log types.Log) (*Event, error) {
    // Check if this is from one of our tracked tokens
    isTracked := false
    for _, token := range p.tokens {
        if log.Address == token {
            isTracked = true
            break
        }
    }
    
    if !isTracked {
        return nil, nil // Skip non-tracked tokens
    }
    
    if len(log.Topics) < 3 {
        return nil, fmt.Errorf("invalid transfer topics")
    }
    
    amount := new(big.Int).SetBytes(log.Data)
    
    event := &Event{
        BlockNumber:     log.BlockNumber,
        BlockHash:       log.BlockHash.Hex(),
        TransactionHash: log.TxHash.Hex(),
        LogIndex:        log.Index,
        Address:         log.Address,
        Protocol:        p.protocolName,
        EventName:       "Transfer",
        Data: EventData{
            "from":   common.BytesToAddress(log.Topics[1].Bytes()),
            "to":     common.BytesToAddress(log.Topics[2].Bytes()),
            "amount": amount.String(),
            "token":  log.Address.Hex(),
        },
    }
    
    return event, nil
}