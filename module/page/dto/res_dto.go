package page_dto

import (
	"time"

	field_dto "auth-login/module/field/dto"
)

const (
	KindContact           = "contact"
	KindEmailVerification = "email_verification"
	KindCustom            = "custom"
)

type Response struct {
	ID           int64     `json:"id"`
	TournamentID int64     `json:"tournament_id"`
	CategoryID   int64     `json:"category_id"`
	Name         string    `json:"name"`
	Sort         int       `json:"sort"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListResponse struct {
	Data []Response `json:"data"`
}

// Fixed steps have no database ID and cannot be modified through CRUD.
type FormPage struct {
	ID           *int64                `json:"id"`
	TournamentID int64                 `json:"tournament_id"`
	CategoryID   int64                 `json:"category_id"`
	Name         string                `json:"name"`
	Sort         int                   `json:"sort"`
	Kind         string                `json:"kind"`
	Fields       []field_dto.FormField `json:"fields"`
}

type FormResponse struct {
	Data []FormPage `json:"data"`
}
