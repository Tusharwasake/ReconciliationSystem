-- Create database tables for ReconciliationSystem

-- Records table to store payment and settlement records
CREATE TABLE IF NOT EXISTS records (
    id SERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL,
    order_id VARCHAR(100) NOT NULL,
    date TIMESTAMP NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    raw_data TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Reconciled records table to store reconciliation results
CREATE TABLE IF NOT EXISTS reconciled_records (
    id SERIAL PRIMARY KEY,
    payments_record_id INTEGER REFERENCES records(id),
    settlements_record_id INTEGER REFERENCES records(id),
    amount_difference DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_records_source ON records(source);
CREATE INDEX IF NOT EXISTS idx_records_order_id ON records(order_id);
CREATE INDEX IF NOT EXISTS idx_records_date ON records(date);
CREATE INDEX IF NOT EXISTS idx_reconciled_payments ON reconciled_records(payments_record_id);
CREATE INDEX IF NOT EXISTS idx_reconciled_settlements ON reconciled_records(settlements_record_id);
