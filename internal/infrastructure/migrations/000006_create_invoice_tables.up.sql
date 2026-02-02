-- 000006_create_invoice_tables.up.sql

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    third_party_id UUID NOT NULL,
    number VARCHAR(100) NOT NULL UNIQUE,
    date TIMESTAMPTZ NOT NULL,
    total_amount NUMERIC(15, 4) NOT NULL,
    total_cost NUMERIC(15, 4) NOT NULL,
    FOREIGN KEY (third_party_id) REFERENCES third_parties(id) ON DELETE RESTRICT,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_invoices_third_party_id ON invoices(third_party_id);
CREATE INDEX idx_invoices_date ON invoices(date);

CREATE TABLE invoice_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    created_by UUID,
    updated_by UUID,
    invoice_id UUID NOT NULL,
    item_id UUID NOT NULL,
    description VARCHAR(255) NOT NULL,
    quantity NUMERIC(15, 4) NOT NULL,
    unit_price NUMERIC(15, 4) NOT NULL,
    unit_cost NUMERIC(15, 4) NOT NULL,
    total_amount NUMERIC(15, 4) NOT NULL,
    total_cost NUMERIC(15, 4) NOT NULL,
    FOREIGN KEY (invoice_id) REFERENCES invoices(id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(id) ON DELETE RESTRICT,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX idx_invoice_lines_invoice_id ON invoice_lines(invoice_id);
CREATE INDEX idx_invoice_lines_item_id ON invoice_lines(item_id);
