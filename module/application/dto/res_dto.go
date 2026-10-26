package application_dto

import "time"

type Response struct {
	ID              int64      `json:"application_id"`
	TournamentID    int64      `json:"tournament_id"`
	CategoryID      int64      `json:"category_id"`
	Name            string     `json:"name"`
	Email           string     `json:"email"`
	BirthDate       string     `json:"birth_date"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateResponse struct {
	Response
	CodeExpiresAt time.Time `json:"code_expires_at"`
}

type Verification struct {
	TournamentID  int64     `json:"tournament_id"`
	CategoryID    int64     `json:"category_id"`
	ApplicationID int64     `json:"application_id"`
	Code          string    `json:"code"`
	ExpiresAt     time.Time `json:"expires_at"`
	Attempts      int       `json:"attempts"`
}

type DataValue struct {
	Field int64  `json:"field"`
	Value string `json:"value"`
}

type DataResponse struct {
	ApplicationID int64       `json:"application_id"`
	Page          string      `json:"page"`
	Data          []DataValue `json:"data"`
	Created       bool        `json:"-"`
}
