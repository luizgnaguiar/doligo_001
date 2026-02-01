CREATE TABLE IF NOT EXISTS bill_of_materials (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT fk_bill_of_materials_product FOREIGN KEY (product_id) REFERENCES items(id) ON DELETE RESTRICT,
    CONSTRAINT fk_bill_of_materials_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_bill_of_materials_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS bill_of_materials_components (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_of_materials_id UUID NOT NULL,
    component_item_id UUID NOT NULL,
    quantity NUMERIC(15,4) NOT NULL,
    unit_of_measure VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID NOT NULL,
    updated_by UUID NOT NULL,
    CONSTRAINT fk_bom_components_bom FOREIGN KEY (bill_of_materials_id) REFERENCES bill_of_materials(id) ON DELETE CASCADE,
    CONSTRAINT fk_bom_components_component_item FOREIGN KEY (component_item_id) REFERENCES items(id) ON DELETE RESTRICT,
    CONSTRAINT fk_bom_components_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_bom_components_updated_by FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE RESTRICT,
    UNIQUE (bill_of_materials_id, component_item_id) -- A component can only be listed once per BOM
);

CREATE TABLE IF NOT EXISTS production_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bill_of_materials_id UUID NOT NULL,
    produced_product_id UUID NOT NULL,
    production_quantity NUMERIC(15,4) NOT NULL,
    actual_production_cost NUMERIC(15,4) NOT NULL,
    warehouse_id UUID NOT NULL,
    produced_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_by UUID NOT NULL,
    CONSTRAINT fk_production_records_bom FOREIGN KEY (bill_of_materials_id) REFERENCES bill_of_materials(id) ON DELETE RESTRICT,
    CONSTRAINT fk_production_records_produced_product FOREIGN KEY (produced_product_id) REFERENCES items(id) ON DELETE RESTRICT,
    CONSTRAINT fk_production_records_warehouse FOREIGN KEY (warehouse_id) REFERENCES warehouses(id) ON DELETE RESTRICT,
    CONSTRAINT fk_production_records_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);
