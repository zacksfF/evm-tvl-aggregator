-- Initial schema for TVL Aggregator
-- This file is executed when the PostgreSQL container starts for the first time

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Chains table
CREATE TABLE chains (
    id SERIAL PRIMARY KEY,
    chain_id INTEGER UNIQUE NOT NULL,
    name VARCHAR(50) NOT NULL,
    rpc_url VARCHAR(255) NOT NULL,
    ws_url VARCHAR(255),
    block_time INTEGER NOT NULL DEFAULT 12,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Protocols table
CREATE TABLE protocols (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_id VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    category VARCHAR(50) NOT NULL,
    website VARCHAR(255),
    description TEXT,
    logo_url VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Protocol deployments (protocol on specific chain)
CREATE TABLE protocol_deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_id UUID REFERENCES protocols(id) ON DELETE CASCADE,
    chain_id INTEGER REFERENCES chains(chain_id) ON DELETE CASCADE,
    contract_address VARCHAR(42),
    deployment_block BIGINT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(protocol_id, chain_id)
);

-- TVL snapshots - stores TVL data at specific timestamps
CREATE TABLE tvl_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_deployment_id UUID REFERENCES protocol_deployments(id) ON DELETE CASCADE,
    tvl_usd DECIMAL(24, 2) NOT NULL,
    token_balances JSONB,
    block_number BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (protocol_deployment_id, timestamp),
    INDEX (timestamp),
    INDEX (block_number)
);

-- Tokens table
CREATE TABLE tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    address VARCHAR(42) NOT NULL,
    chain_id INTEGER REFERENCES chains(chain_id) ON DELETE CASCADE,
    symbol VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    decimals INTEGER NOT NULL,
    is_stable BOOLEAN DEFAULT false,
    coingecko_id VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(address, chain_id)
);

-- Token prices table
CREATE TABLE token_prices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_id UUID REFERENCES tokens(id) ON DELETE CASCADE,
    price_usd DECIMAL(24, 8) NOT NULL,
    source VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (token_id, timestamp),
    INDEX (timestamp)
);

-- Protocol token holdings
CREATE TABLE protocol_token_holdings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_deployment_id UUID REFERENCES protocol_deployments(id) ON DELETE CASCADE,
    token_id UUID REFERENCES tokens(id) ON DELETE CASCADE,
    balance DECIMAL(36, 18) NOT NULL,
    value_usd DECIMAL(24, 2) NOT NULL,
    percentage DECIMAL(5, 2),
    block_number BIGINT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (protocol_deployment_id, timestamp),
    INDEX (token_id, timestamp)
);

-- Events table for tracking blockchain events
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id INTEGER REFERENCES chains(chain_id) ON DELETE CASCADE,
    contract_address VARCHAR(42) NOT NULL,
    event_name VARCHAR(100) NOT NULL,
    event_signature VARCHAR(66) NOT NULL,
    block_number BIGINT NOT NULL,
    transaction_hash VARCHAR(66) NOT NULL,
    log_index INTEGER NOT NULL,
    event_data JSONB,
    processed BOOLEAN DEFAULT false,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (chain_id, block_number),
    INDEX (contract_address, event_name),
    INDEX (processed),
    UNIQUE(transaction_hash, log_index)
);

-- Indexer checkpoints
CREATE TABLE indexer_checkpoints (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id INTEGER REFERENCES chains(chain_id) ON DELETE CASCADE,
    contract_address VARCHAR(42),
    last_processed_block BIGINT NOT NULL,
    last_processed_timestamp TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(chain_id, contract_address)
);

-- Aggregated TVL by chain
CREATE TABLE chain_tvl_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chain_id INTEGER REFERENCES chains(chain_id) ON DELETE CASCADE,
    total_tvl_usd DECIMAL(24, 2) NOT NULL,
    protocol_count INTEGER NOT NULL DEFAULT 0,
    daily_change_percentage DECIMAL(8, 4),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (chain_id, timestamp),
    INDEX (timestamp)
);

-- Aggregated TVL by protocol
CREATE TABLE protocol_tvl_aggregates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    protocol_id UUID REFERENCES protocols(id) ON DELETE CASCADE,
    total_tvl_usd DECIMAL(24, 2) NOT NULL,
    chain_count INTEGER NOT NULL DEFAULT 0,
    daily_change_percentage DECIMAL(8, 4),
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    INDEX (protocol_id, timestamp),
    INDEX (timestamp)
);

-- Create indexes for better performance
CREATE INDEX idx_tvl_snapshots_timestamp ON tvl_snapshots(timestamp DESC);
CREATE INDEX idx_token_prices_timestamp ON token_prices(timestamp DESC);
CREATE INDEX idx_events_block_number ON events(chain_id, block_number);
CREATE INDEX idx_protocol_deployments_active ON protocol_deployments(is_active) WHERE is_active = true;

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add updated_at triggers
CREATE TRIGGER update_chains_updated_at BEFORE UPDATE ON chains FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_protocols_updated_at BEFORE UPDATE ON protocols FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_protocol_deployments_updated_at BEFORE UPDATE ON protocol_deployments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tokens_updated_at BEFORE UPDATE ON tokens FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_indexer_checkpoints_updated_at BEFORE UPDATE ON indexer_checkpoints FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();