-- Create trades table
CREATE TABLE IF NOT EXISTS trades (
    event_type String NOT NULL,
    event_time Int64 NOT NULL,
    symbol LowCardinality(String) NOT NULL,
    trade_id Int64 NOT NULL,
    price Decimal(18, 8) NOT NULL,
    quantity Float64 NOT NULL,
    trade_time DateTime64(3) NOT NULL,
    is_market_maker Bool NOT NULL
) ENGINE = MergeTree()
ORDER BY (symbol, trade_time, trade_id)

-- Modify quantity column to Decimal
ALTER TABLE trades MODIFY COLUMN quantity Decimal(18, 8)