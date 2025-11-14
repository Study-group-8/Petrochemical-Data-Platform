-- Initialize PostgreSQL database for Petrochemical Data Platform

-- Create assets table
CREATE TABLE IF NOT EXISTS assets (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    location VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on assets
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(type);
CREATE INDEX IF NOT EXISTS idx_assets_location ON assets(location);

-- Create control_commands table
CREATE TABLE IF NOT EXISTS control_commands (
    id VARCHAR(255) PRIMARY KEY,
    equipment_id VARCHAR(255) NOT NULL,
    command VARCHAR(255) NOT NULL,
    parameters JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    executed_at TIMESTAMP WITH TIME ZONE
);

-- Create index on control commands
CREATE INDEX IF NOT EXISTS idx_control_commands_equipment_id ON control_commands(equipment_id);
CREATE INDEX IF NOT EXISTS idx_control_commands_status ON control_commands(status);

-- Insert sample data
INSERT INTO assets (id, name, type, location) VALUES
('ROSNEFT_OMSK_REFINERY', 'Роснефть - Омский НПЗ', 'refinery', 'Омск'),
('GAZPROMNEFT_MOSCOW_REFINERY', 'Газпромнефть - Московский НПЗ', 'refinery', 'Москва'),
('LUKOIL_VOLGOGRAD_REFINERY', 'Лукойл - Волгоградский НПЗ', 'refinery', 'Волгоград'),
('SIBUR_TOBOLSK_POLYMER', 'СИБУР - Тобольск Полимер', 'polymer_plant', 'Тобольск'),
('TATNEFT_ROMASHKINO_FIELD', 'Татнефть - Ромашкинское месторождение', 'oilfield', 'Ромашкино'),
('NOVATEK_YAMAL_LNG', 'Новатэк - Ямал СПГ', 'lng_plant', 'Ямал')
ON CONFLICT (id) DO NOTHING;