package tournament_dto

import "time"

type Create struct {
	Name        string    `json:"name" validate:"required,max=200"`
	Description string    `json:"description" validate:"max=10000"`
	Location    string    `json:"location" validate:"max=300"`
	StartsAt    time.Time `json:"starts_at" validate:"required"`
	EndsAt      time.Time `json:"ends_at" validate:"required,gtfield=StartsAt"`
	Status      string    `json:"status" validate:"required,oneof=planned ongoing completed cancelled"`
}

type Update Create

type Filter struct {
	Cursor      int64
	Limit       int
	Name        string
	Description string
	Status      string
	Direction   string
}
