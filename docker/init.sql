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

-- Create sensors table
CREATE TABLE IF NOT EXISTS sensors (
    id VARCHAR(255) PRIMARY KEY,
    asset_id VARCHAR(255) REFERENCES assets(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(100) NOT NULL,
    unit VARCHAR(50),
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on sensors
CREATE INDEX IF NOT EXISTS idx_sensors_asset_id ON sensors(asset_id);
CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(type);

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

INSERT INTO sensors (id, asset_id, name, type, unit) VALUES
('ROSNEFT_REFINERY_01_TEMP_001', 'ROSNEFT_OMSK_REFINERY', 'Реактор гидрокрекинга T-101', 'temperature', '°C'),
('GAZPROMNEFT_REFINERY_02_PRESS_001', 'GAZPROMNEFT_MOSCOW_REFINERY', 'Колонна фракционирования P-201', 'pressure', 'бар'),
('LUKOIL_REFINERY_03_LEVEL_001', 'LUKOIL_VOLGOGRAD_REFINERY', 'Резервуар сырья L-301', 'level', '%'),
('SIBUR_POLYMER_01_FLOW_001', 'SIBUR_TOBOLSK_POLYMER', 'Линия полипропилена F-401', 'flow', 'м³/ч'),
('TATNEFT_OILFIELD_01_VIBR_001', 'TATNEFT_ROMASHKINO_FIELD', 'Насосная станция V-501', 'vibration', 'мм/с'),
('NOVATEK_LNG_01_DENSITY_001', 'NOVATEK_YAMAL_LNG', 'Хранилище сжиженного газа D-601', 'density', 'кг/м³')
ON CONFLICT (id) DO NOTHING;