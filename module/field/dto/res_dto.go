package field_dto

import "time"

const (
	TypeText     = "text"
	TypeTextarea = "textarea"
	TypeEmail    = "email"
	TypeTel      = "tel"
	TypeNumber   = "number"
	TypeDate     = "date"
	TypeSelect   = "select"
	TypeRadio    = "radio"
	TypeCheckbox = "checkbox"
	TypeFile     = "file"
)

type Response struct {
	ID         int64      `json:"id"`
	PageID     int64      `json:"page_id"`
	Sort       int        `json:"sort"`
	Type       string     `json:"type"`
	Options    []string   `json:"options"`
	Name       string     `json:"name"`
	Validation Validation `json:"validation"`
	IsActive   bool       `json:"is_active"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ListResponse struct {
	Data []Response `json:"data"`
}

type FormField struct {
	ID         *int64     `json:"id"`
	Type       string     `json:"type"`
	Name       string     `json:"name"`
	Validation Validation `json:"validation"`
	Options    []string   `json:"options"`
}
