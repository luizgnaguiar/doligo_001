-- 000003_create_thirdparties_and_items_tables.up.sql
-- This script creates the third_parties and items tables.

-- Create the third_parties table for customers and suppliers
CREATE TABLE IF NOT EXISTS third_parties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'CUSTOMER' or 'SUPPLIER'
    is_active BOOLEAN DEFAULT TRUE
);

-- Create the items table for products and services
CREATE TABLE IF NOT EXISTS items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    updated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'STORABLE' or 'SERVICE'
    cost_price NUMERIC(15, 4) DEFAULT 0.0,
    sale_price NUMERIC(15, 4) DEFAULT 0.0,
    average_cost NUMERIC(15, 4) DEFAULT 0.0,
    is_active BOOLEAN DEFAULT TRUE
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_third_parties_name ON third_parties(name);
CREATE INDEX IF NOT EXISTS idx_items_name ON items(name);
