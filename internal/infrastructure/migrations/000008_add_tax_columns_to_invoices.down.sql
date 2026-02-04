ALTER TABLE invoice_lines DROP COLUMN tax_rate;
ALTER TABLE invoice_lines DROP COLUMN tax_amount;
ALTER TABLE invoice_lines DROP COLUMN net_price;
ALTER TABLE invoices DROP COLUMN total_tax;
