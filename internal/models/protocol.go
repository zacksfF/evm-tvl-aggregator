package models

import (
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// Protocol represents a DeFi protocol
type Protocol struct {
	ID          uint64                      `json:"id" db:"id"`
	Name        string                      `json:"name" db:"name"`
	Type        ProtocolType                `json:"type" db:"type"`
	Description string                      `json:"description" db:"description"`
	Website     string                      `json:"website" db:"website"`
	Logo        string                      `json:"logo" db:"logo"`
	Chains      map[string][]ContractConfig `json:"chains"`
	CreatedAt   time.Time                   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at" db:"updated_at"`
}

// ProtocolType defines the type of protocol
type ProtocolType string

const (
	ProtocolTypeDEX        ProtocolType = "dex"
	ProtocolTypeLending    ProtocolType = "lending"
	ProtocolTypeYield      ProtocolType = "yield"
	ProtocolTypeBridge     ProtocolType = "bridge"
	ProtocolTypeDerivative ProtocolType = "derivative"
	ProtocolTypeInsurance  ProtocolType = "insurance"
)

// ContractConfig represents a smart contract configuration
type ContractConfig struct {
	Address     common.Address   `json:"address"`
	Name        string           `json:"name"`
	Type        string           `json:"type"`
	Version     string           `json:"version"`
	Tokens      []common.Address `json:"tokens,omitempty"`
	PoolID      string           `json:"pool_id,omitempty"`
	VaultID     string           `json:"vault_id,omitempty"`
	DeployBlock uint64           `json:"deploy_block,omitempty"`
	ABI         string           `json:"abi,omitempty"`
}
