-- Seed data for TVL Aggregator
-- Insert initial chain configurations

INSERT INTO chains (chain_id, name, rpc_url, ws_url, block_time, is_active) VALUES
(1, 'Ethereum', 'https://eth-mainnet.g.alchemy.com/v2/YOUR_ALCHEMY_KEY', 'wss://eth-mainnet.g.alchemy.com/v2/YOUR_ALCHEMY_KEY', 12, true),
(56, 'BSC', 'https://bsc-dataseed1.binance.org/', 'wss://bsc-ws-node.nariox.org:443', 3, true),
(137, 'Polygon', 'https://polygon-rpc.com/', 'wss://ws-mainnet.matic.network', 2, true),
(42161, 'Arbitrum One', 'https://arb1.arbitrum.io/rpc', 'wss://arb1.arbitrum.io/ws', 1, true),
(10, 'Optimism', 'https://mainnet.optimism.io/', 'wss://ws-mainnet.optimism.io', 2, true),
(43114, 'Avalanche', 'https://api.avax.network/ext/bc/C/rpc', 'wss://api.avax.network/ext/bc/C/ws', 2, true),
(8453, 'Base', 'https://mainnet.base.org/', 'wss://mainnet.base.org/ws', 2, true),
(250, 'Fantom', 'https://rpc.ftm.tools/', 'wss://ws.fantom.network', 1, true);

-- Insert major DeFi protocols
INSERT INTO protocols (protocol_id, name, category, website, description) VALUES
('aave-v3', 'Aave V3', 'Lending', 'https://aave.com', 'Decentralized lending and borrowing protocol'),
('uniswap-v3', 'Uniswap V3', 'DEX', 'https://uniswap.org', 'Automated liquidity protocol'),
('compound-v3', 'Compound V3', 'Lending', 'https://compound.finance', 'Autonomous interest rate protocol'),
('curve-finance', 'Curve Finance', 'DEX', 'https://curve.fi', 'Exchange liquidity pool for stablecoins'),
('makerdao', 'MakerDAO', 'CDP', 'https://makerdao.com', 'Decentralized autonomous organization'),
('lido', 'Lido', 'Staking', 'https://lido.fi', 'Liquid staking for Ethereum'),
('convex-finance', 'Convex Finance', 'Yield', 'https://convexfinance.com', 'Platform for CRV and CVX holders'),
('balancer', 'Balancer', 'DEX', 'https://balancer.fi', 'Automated portfolio manager and liquidity provider'),
('yearn-finance', 'Yearn Finance', 'Yield', 'https://yearn.finance', 'Yield farming protocol'),
('sushiswap', 'SushiSwap', 'DEX', 'https://sushi.com', 'Community-driven DEX'),
('pancakeswap', 'PancakeSwap', 'DEX', 'https://pancakeswap.finance', 'Leading DEX on BSC'),
('quickswap', 'QuickSwap', 'DEX', 'https://quickswap.exchange', 'Next-gen DEX on Polygon'),
('trader-joe', 'Trader Joe', 'DEX', 'https://traderjoexyz.com', 'One-stop decentralized trading'),
('gmx', 'GMX', 'Derivatives', 'https://gmx.io', 'Decentralized perpetual exchange'),
('radiant-capital', 'Radiant Capital', 'Lending', 'https://radiant.capital', 'Cross-chain lending protocol');

-- Insert protocol deployments for major chains
-- Aave V3 deployments
INSERT INTO protocol_deployments (protocol_id, chain_id, contract_address, deployment_block) VALUES
((SELECT id FROM protocols WHERE protocol_id = 'aave-v3'), 1, '0x87870Bca3F3fD6335C3F4ce8392D69350B4fA4E2', 16291127),
((SELECT id FROM protocols WHERE protocol_id = 'aave-v3'), 137, '0x794a61358D6845594F94dc1DB02A252b5b4814aD', 25825996),
((SELECT id FROM protocols WHERE protocol_id = 'aave-v3'), 42161, '0x794a61358D6845594F94dc1DB02A252b5b4814aD', 7740429),
((SELECT id FROM protocols WHERE protocol_id = 'aave-v3'), 10, '0x794a61358D6845594F94dc1DB02A252b5b4814aD', 4365693),
((SELECT id FROM protocols WHERE protocol_id = 'aave-v3'), 43114, '0x794a61358D6845594F94dc1DB02A252b5b4814aD', 11970654);

-- Uniswap V3 deployments
INSERT INTO protocol_deployments (protocol_id, chain_id, contract_address, deployment_block) VALUES
((SELECT id FROM protocols WHERE protocol_id = 'uniswap-v3'), 1, '0x1F98431c8aD98523631AE4a59f267346ea31F984', 12369621),
((SELECT id FROM protocols WHERE protocol_id = 'uniswap-v3'), 137, '0x1F98431c8aD98523631AE4a59f267346ea31F984', 15259656),
((SELECT id FROM protocols WHERE protocol_id = 'uniswap-v3'), 42161, '0x1F98431c8aD98523631AE4a59f267346ea31F984', 165),
((SELECT id FROM protocols WHERE protocol_id = 'uniswap-v3'), 10, '0x1F98431c8aD98523631AE4a59f267346ea31F984', 4369710),
((SELECT id FROM protocols WHERE protocol_id = 'uniswap-v3'), 8453, '0x33128a8fC17869897dcE68Ed026d694621f6FDfD', 1371680);

-- Compound V3 deployments
INSERT INTO protocol_deployments (protocol_id, chain_id, contract_address, deployment_block) VALUES
((SELECT id FROM protocols WHERE protocol_id = 'compound-v3'), 1, '0xc3d688B66703497DAA19211EEdff47f25384cdc3', 15331586),
((SELECT id FROM protocols WHERE protocol_id = 'compound-v3'), 137, '0xF25212E676D1F7F89Cd72fFEe66158f541246445', 30944675),
((SELECT id FROM protocols WHERE protocol_id = 'compound-v3'), 42161, '0xA5EDBDD9646f8dFF606d7448e414884C7d905dCA', 57469158);

-- Insert major tokens
INSERT INTO tokens (address, chain_id, symbol, name, decimals, is_stable, coingecko_id) VALUES
-- Ethereum tokens
('0x0000000000000000000000000000000000000000', 1, 'ETH', 'Ethereum', 18, false, 'ethereum'),
('0xA0b86a33E6417c8C4c2c6c1E0B1F4c8b1C8b9B1e', 1, 'USDC', 'USD Coin', 6, true, 'usd-coin'),
('0x6B175474E89094C44Da98b954EedeAC495271d0F', 1, 'DAI', 'Dai Stablecoin', 18, true, 'dai'),
('0xdAC17F958D2ee523a2206206994597C13D831ec7', 1, 'USDT', 'Tether USD', 6, true, 'tether'),
('0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599', 1, 'WBTC', 'Wrapped BTC', 8, false, 'wrapped-bitcoin'),
('0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2', 1, 'WETH', 'Wrapped Ether', 18, false, 'weth'),

-- Polygon tokens
('0x0000000000000000000000000000000000001010', 137, 'MATIC', 'Polygon', 18, false, 'matic-network'),
('0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174', 137, 'USDC', 'USD Coin (PoS)', 6, true, 'usd-coin'),
('0x8f3Cf7ad23Cd3CaDbD9735AFf958023239c6A063', 137, 'DAI', 'Dai Stablecoin (PoS)', 18, true, 'dai'),
('0xc2132D05D31c914a87C6611C10748AEb04B58e8F', 137, 'USDT', 'Tether USD (PoS)', 6, true, 'tether'),

-- BSC tokens
('0x0000000000000000000000000000000000000000', 56, 'BNB', 'BNB', 18, false, 'binancecoin'),
('0x8AC76a51cc950d9822D68b83fE1Ad97B32Cd580d', 56, 'USDC', 'USD Coin', 18, true, 'usd-coin'),
('0x55d398326f99059fF775485246999027B3197955', 56, 'USDT', 'Tether USD', 18, true, 'tether'),
('0xe9e7CEA3DedcA5984780Bafc599bD69ADd087D56', 56, 'BUSD', 'BUSD Token', 18, true, 'binance-usd'),

-- Arbitrum tokens
('0x0000000000000000000000000000000000000000', 42161, 'ETH', 'Ethereum', 18, false, 'ethereum'),
('0xFF970A61A04b1cA14834A43f5dE4533eBDDB5CC8', 42161, 'USDC', 'USD Coin (Arb1)', 6, true, 'usd-coin'),
('0xDA10009cBd5D07dd0CeCc66161FC93D7c9000da1', 42161, 'DAI', 'Dai Stablecoin', 18, true, 'dai'),
('0xFd086bC7CD5C481DCC9C85ebE478A1C0b69FCbb9', 42161, 'USDT', 'Tether USD', 6, true, 'tether');

-- Insert initial indexer checkpoints (starting from recent blocks)
INSERT INTO indexer_checkpoints (chain_id, contract_address, last_processed_block, last_processed_timestamp) VALUES
(1, NULL, 19500000, NOW() - INTERVAL '1 hour'),
(56, NULL, 36000000, NOW() - INTERVAL '1 hour'),
(137, NULL, 54000000, NOW() - INTERVAL '1 hour'),
(42161, NULL, 180000000, NOW() - INTERVAL '1 hour'),
(10, NULL, 115000000, NOW() - INTERVAL '1 hour'),
(43114, NULL, 40000000, NOW() - INTERVAL '1 hour');

COMMIT;