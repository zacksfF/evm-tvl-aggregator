package postgres

const migrationSQL = `
-- Protocols table
CREATE TABLE IF NOT EXISTS protocols (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL,
    description TEXT,
    website VARCHAR(255),
    logo VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Chains table
CREATE TABLE IF NOT EXISTS chains (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    chain_id BIGINT UNIQUE NOT NULL,
    rpc_endpoint TEXT,
    ws_endpoint TEXT,
    explorer VARCHAR(255),
    native_token VARCHAR(10),
    block_time INT DEFAULT 12,
    is_testnet BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Contracts table
CREATE TABLE IF NOT EXISTS contracts (
    id SERIAL PRIMARY KEY,
    protocol_id INT REFERENCES protocols(id) ON DELETE CASCADE,
    chain VARCHAR(50),
    address VARCHAR(42) NOT NULL,
    name VARCHAR(100),
    type VARCHAR(50),
    version VARCHAR(20),
    tokens JSONB,
    pool_id VARCHAR(100),
    vault_id VARCHAR(100),
    deploy_block BIGINT,
    abi TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain, address)
);

-- TVL snapshots table
CREATE TABLE IF NOT EXISTS tvl_snapshots (
    id SERIAL PRIMARY KEY,
    protocol VARCHAR(100) NOT NULL,
    chain VARCHAR(50),
    block_number BIGINT,
    total_usd NUMERIC(30, 8),
    breakdown JSONB,
    timestamp TIMESTAMP NOT NULL,
    UNIQUE(protocol, chain, block_number)
);

-- Events table
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    chain VARCHAR(50) NOT NULL,
    protocol VARCHAR(100),
    block_number BIGINT NOT NULL,
    block_hash VARCHAR(66),
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INT NOT NULL,
    address VARCHAR(42) NOT NULL,
    event_name VARCHAR(100),
    event_signature VARCHAR(100),
    data JSONB,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain, transaction_hash, log_index)
);

-- Indexed blocks table
CREATE TABLE IF NOT EXISTS indexed_blocks (
    id SERIAL PRIMARY KEY,
    chain VARCHAR(50) NOT NULL,
    number BIGINT NOT NULL,
    hash VARCHAR(66),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain, number)
);

-- Tokens table
CREATE TABLE IF NOT EXISTS tokens (
    id SERIAL PRIMARY KEY,
    address VARCHAR(42) NOT NULL,
    chain VARCHAR(50) NOT NULL,
    symbol VARCHAR(20),
    name VARCHAR(100),
    decimals INT DEFAULT 18,
    total_supply NUMERIC(78, 0),
    logo VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(chain, address)
);

-- Token prices table
CREATE TABLE IF NOT EXISTS token_prices (
    id SERIAL PRIMARY KEY,
    token_id INT REFERENCES tokens(id),
    price_usd NUMERIC(30, 18),
    source VARCHAR(50),
    confidence FLOAT DEFAULT 1.0,
    timestamp TIMESTAMP NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_tvl_protocol_time ON tvl_snapshots(protocol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_tvl_chain_time ON tvl_snapshots(chain, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_events_chain_block ON events(chain, block_number);
CREATE INDEX IF NOT EXISTS idx_events_protocol ON events(protocol);
CREATE INDEX IF NOT EXISTS idx_token_prices_time ON token_prices(token_id, timestamp DESC);
`

func (ps *PostgresStorage) migrate() error {
	_, err := ps.db.Exec(migrationSQL)
	return err
}
