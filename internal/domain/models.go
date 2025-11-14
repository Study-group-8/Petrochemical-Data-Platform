package domain

import "time"

// Product представляет продукт компании (бывший "датчик")
type Product struct {
	ID        string    `json:"id"`
	CompanyID string    `json:"company_id"`
	Name      string    `json:"name"`   // Полипропилен, Полиэтилен и т.д.
	Type      string    `json:"type"`   // polymer, fuel, chemical
	Unit      string    `json:"unit"`   // т/час, м³/час
	Status    string    `json:"status"` // active, inactive
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TelemetryData представляет данные о продажах/производстве продуктов
type TelemetryData struct {
	CompanyID   string    `json:"company_id"`   // ID компании
	ProductName string    `json:"product_name"` // Название продукта
	Value       float64   `json:"value"`        // Объём производства
	Unit        string    `json:"unit"`         // Единица измерения
	Timestamp   time.Time `json:"timestamp"`
	Quality     uint16    `json:"quality"` // 0-bad, 1-good
	Tags        []string  `json:"tags,omitempty"`
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

// Asset представляет промышленное оборудование или компанию
type Asset struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Type           string                 `json:"type"`     // company, refinery, plant
	Location       string                 `json:"location"` // Регион
	Status         string                 `json:"status"`
	Specifications map[string]interface{} `json:"specifications,omitempty"`
	Products       []string               `json:"products,omitempty"` // Список продуктов компании
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Alert представляет системные оповещения и уведомления
type Alert struct {
	ID        string     `json:"id"`
	Type      string     `json:"type"`
	Severity  string     `json:"severity"`
	Message   string     `json:"message"`
	CompanyID string     `json:"company_id,omitempty"` // Заменили sensor_id на company_id
	AssetID   string     `json:"asset_id,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
	Acked     bool       `json:"acked"`
	AckedAt   *time.Time `json:"acked_at,omitempty"`
}
