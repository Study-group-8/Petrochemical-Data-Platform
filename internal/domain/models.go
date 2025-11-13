package domain

import "time"

// Sensor представляет устройство датчика
type Sensor struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Location  string                 `json:"location"`
	Status    string                 `json:"status"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// TelemetryData представляет данные временных рядов от датчиков
type TelemetryData struct {
	SensorID  string    `json:"sensor_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Timestamp time.Time `json:"timestamp"`
	Quality   uint16    `json:"quality"`
	Tags      []string  `json:"tags,omitempty"`
}

// ControlCommand представляет команду управления оборудованием
type ControlCommand struct {
	ID          string                 `json:"id"`
	EquipmentID string                 `json:"equipment_id"`
	Command     string                 `json:"command"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	ExecutedAt  *time.Time             `json:"executed_at,omitempty"`
}

// Asset представляет промышленное оборудование или актив
type Asset struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`
	Location       string                 `json:"location"`
	Status         string                 `json:"status"`
	Specifications map[string]interface{} `json:"specifications,omitempty"`
	Sensors        []string               `json:"sensors,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Alert представляет системные оповещения и уведомления
type Alert struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Severity  string     `json:"severity"`
	Message   string     `json:"message"`
	SensorID  string     `json:"sensor_id,omitempty"`
	AssetID   string     `json:"asset_id,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Acked     bool       `json:"acked"`
	AckedAt   *time.Time `json:"acked_at,omitempty"`
}
