ALTER TABLE invoices ADD COLUMN pdf_status VARCHAR(20) DEFAULT 'pending';
ALTER TABLE invoices ADD COLUMN pdf_url TEXT;
