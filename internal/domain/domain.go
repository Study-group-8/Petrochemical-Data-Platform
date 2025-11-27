package domain

import "time"

// Asset represents equipment metadata used across services/repositories
type Asset struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"`
	Location  string    `json:"location" db:"location"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// TelemetryData represents a telemetry record
type TelemetryData struct {
	CompanyID   string    `json:"company_id"`
	ProductName string    `json:"product_name"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	Quality     uint16    `json:"quality"`
}

// ControlCommand represents a control action
type ControlCommand struct {
	ID          string `json:"id"`
	EquipmentID string `json:"equipment_id"`
	Command     string `json:"command"`
}
