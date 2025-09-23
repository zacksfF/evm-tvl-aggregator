// internal/storage/postgres/postgres.go
package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/zacksfF/evm-tvl-aggregator/internal/models"
	"github.com/zacksfF/evm-tvl-aggregator/internal/storage"
)

// PostgresStorage implements the Storage interface
type PostgresStorage struct {
	db *sqlx.DB
}

// PostgresTx implements the Tx interface for PostgreSQL
type PostgresTx struct {
	tx *sqlx.Tx
}

// Commit commits the transaction
func (ptx *PostgresTx) Commit() error {
	return ptx.tx.Commit()
}

// Rollback rolls back the transaction
func (ptx *PostgresTx) Rollback() error {
	return ptx.tx.Rollback()
}

// All other interface methods should be implemented by delegating to PostgresStorage
// with the transaction context. For brevity, I'll add minimal implementations.
func (ptx *PostgresTx) BeginTx(ctx context.Context) (storage.Tx, error) {
	// Can't nest transactions in PostgreSQL in this simple implementation
	return ptx, nil
}

func (ptx *PostgresTx) Close() error {
	return ptx.tx.Rollback() // Rollback if not committed
}

// Protocol methods - these should use the transaction
func (ptx *PostgresTx) GetProtocol(ctx context.Context, name string) (*models.Protocol, error) {
	// TODO: Implement with transaction
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetProtocols(ctx context.Context) ([]*models.Protocol, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) SaveProtocol(ctx context.Context, protocol *models.Protocol) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) UpdateProtocol(ctx context.Context, protocol *models.Protocol) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) DeleteProtocol(ctx context.Context, name string) error {
	return fmt.Errorf("not implemented")
}

// TVL methods
func (ptx *PostgresTx) SaveTVLSnapshot(ctx context.Context, snapshot *models.TVLSnapshot) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetLatestTVL(ctx context.Context, protocol, chain string) (*models.TVLSnapshot, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetHistoricalTVL(ctx context.Context, protocol, chain string, from, to time.Time) ([]*models.TVLSnapshot, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetTVLByBlock(ctx context.Context, protocol, chain string, blockNumber uint64) (*models.TVLSnapshot, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetAggregatedTVL(ctx context.Context) (*models.AggregatedTVL, error) {
	return nil, fmt.Errorf("not implemented")
}

// Chain methods
func (ptx *PostgresTx) GetChain(ctx context.Context, name string) (*models.Chain, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetChains(ctx context.Context) ([]*models.Chain, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) SaveChain(ctx context.Context, chain *models.Chain) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) UpdateChainStats(ctx context.Context, stats *models.ChainStats) error {
	return fmt.Errorf("not implemented")
}

// Event methods
func (ptx *PostgresTx) SaveEvents(ctx context.Context, events []*models.Event) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetEvents(ctx context.Context, filter storage.EventFilter) ([]*models.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) SaveBlock(ctx context.Context, block *models.Block) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetLastIndexedBlock(ctx context.Context, chain string) (uint64, error) {
	return 0, fmt.Errorf("not implemented")
}

// Token methods
func (ptx *PostgresTx) GetToken(ctx context.Context, address, chain string) (*models.Token, error) {
	return nil, fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) SaveToken(ctx context.Context, token *models.Token) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) UpdateTokenPrice(ctx context.Context, price *models.TokenPrice) error {
	return fmt.Errorf("not implemented")
}

func (ptx *PostgresTx) GetTokenPrices(ctx context.Context, addresses []string) (map[string]*models.TokenPrice, error) {
	return nil, fmt.Errorf("not implemented")
}

// NewPostgresStorage creates a new PostgreSQL storage
func NewPostgresStorage(connectionString string) (*PostgresStorage, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ps := &PostgresStorage{db: db}

	// Run migrations
	if err := ps.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return ps, nil
}

// BeginTx starts a new transaction
func (ps *PostgresStorage) BeginTx(ctx context.Context) (storage.Tx, error) {
	tx, err := ps.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &PostgresTx{tx: tx}, nil
}

// Close closes the database connection
func (ps *PostgresStorage) Close() error {
	return ps.db.Close()
}

// GetProtocol retrieves a protocol by name
func (ps *PostgresStorage) GetProtocol(ctx context.Context, name string) (*models.Protocol, error) {
	var protocol models.Protocol
	query := `
        SELECT id, name, type, description, website, logo, created_at, updated_at
        FROM protocols
        WHERE name = $1
    `

	err := ps.db.GetContext(ctx, &protocol, query, name)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("protocol not found: %s", name)
	}
	if err != nil {
		return nil, err
	}

	// Load contracts
	contracts, err := ps.getProtocolContracts(ctx, protocol.ID)
	if err != nil {
		return nil, err
	}
	protocol.Chains = contracts

	return &protocol, nil
}

// GetProtocols retrieves all protocols
func (ps *PostgresStorage) GetProtocols(ctx context.Context) ([]*models.Protocol, error) {
	query := `
        SELECT id, name, type, description, website, logo, created_at, updated_at
        FROM protocols
        WHERE is_active = true
        ORDER BY name
    `

	var protocols []*models.Protocol
	err := ps.db.SelectContext(ctx, &protocols, query)
	if err != nil {
		return nil, err
	}

	// Load contracts for each protocol
	for _, protocol := range protocols {
		contracts, err := ps.getProtocolContracts(ctx, protocol.ID)
		if err != nil {
			continue
		}
		protocol.Chains = contracts
	}

	return protocols, nil
}

// SaveProtocol saves a new protocol
func (ps *PostgresStorage) SaveProtocol(ctx context.Context, protocol *models.Protocol) error {
	tx, err := ps.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
        INSERT INTO protocols (name, type, description, website, logo)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `

	err = tx.GetContext(ctx, &protocol.ID, query,
		protocol.Name, protocol.Type, protocol.Description, protocol.Website, protocol.Logo)
	if err != nil {
		return err
	}

	// Save contracts
	for chain, contracts := range protocol.Chains {
		for _, contract := range contracts {
			if err := ps.saveContractTx(ctx, tx, protocol.ID, chain, contract); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// SaveTVLSnapshot saves a TVL snapshot
func (ps *PostgresStorage) SaveTVLSnapshot(ctx context.Context, snapshot *models.TVLSnapshot) error {
	breakdown, err := json.Marshal(snapshot.Breakdown)
	if err != nil {
		return err
	}

	query := `
        INSERT INTO tvl_snapshots (protocol, chain, block_number, total_usd, breakdown, timestamp)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (protocol, chain, block_number) 
        DO UPDATE SET total_usd = EXCLUDED.total_usd, breakdown = EXCLUDED.breakdown
    `

	totalUSD, _ := snapshot.TotalUSD.Float64()

	_, err = ps.db.ExecContext(ctx, query,
		snapshot.Protocol,
		snapshot.Chain,
		snapshot.BlockNumber,
		totalUSD,
		breakdown,
		snapshot.Timestamp,
	)

	return err
}

// GetLatestTVL retrieves the latest TVL snapshot
func (ps *PostgresStorage) GetLatestTVL(ctx context.Context, protocol, chain string) (*models.TVLSnapshot, error) {
	var snapshot models.TVLSnapshot
	var breakdown []byte
	var totalUSD float64

	query := `
        SELECT id, protocol, chain, block_number, total_usd, breakdown, timestamp
        FROM tvl_snapshots
        WHERE protocol = $1 AND ($2 = '' OR chain = $2)
        ORDER BY timestamp DESC
        LIMIT 1
    `

	err := ps.db.QueryRowContext(ctx, query, protocol, chain).Scan(
		&snapshot.ID,
		&snapshot.Protocol,
		&snapshot.Chain,
		&snapshot.BlockNumber,
		&totalUSD,
		&breakdown,
		&snapshot.Timestamp,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	snapshot.TotalUSD = new(big.Float).SetFloat64(totalUSD)

	if err := json.Unmarshal(breakdown, &snapshot.Breakdown); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

// Helper methods
func (ps *PostgresStorage) getProtocolContracts(ctx context.Context, protocolID uint64) (map[string][]models.ContractConfig, error) {
	query := `
        SELECT chain, address, name, type, version, tokens, pool_id, vault_id, deploy_block
        FROM contracts
        WHERE protocol_id = $1
    `

	rows, err := ps.db.QueryContext(ctx, query, protocolID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contracts := make(map[string][]models.ContractConfig)

	for rows.Next() {
		var chain string
		var contract models.ContractConfig
		var tokens []byte

		err := rows.Scan(
			&chain,
			&contract.Address,
			&contract.Name,
			&contract.Type,
			&contract.Version,
			&tokens,
			&contract.PoolID,
			&contract.VaultID,
			&contract.DeployBlock,
		)
		if err != nil {
			continue
		}

		if tokens != nil {
			json.Unmarshal(tokens, &contract.Tokens)
		}

		contracts[chain] = append(contracts[chain], contract)
	}

	return contracts, nil
}

func (ps *PostgresStorage) saveContractTx(ctx context.Context, tx *sqlx.Tx, protocolID uint64, chain string, contract models.ContractConfig) error {
	tokens, _ := json.Marshal(contract.Tokens)

	query := `
        INSERT INTO contracts (protocol_id, chain, address, name, type, version, tokens, pool_id, vault_id, deploy_block)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	_, err := tx.ExecContext(ctx, query,
		protocolID, chain, contract.Address.Hex(), contract.Name, contract.Type,
		contract.Version, tokens, contract.PoolID, contract.VaultID, contract.DeployBlock,
	)

	return err
}

// Additional methods implementation...
