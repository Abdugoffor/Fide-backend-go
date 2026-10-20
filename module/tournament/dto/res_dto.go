package tournament_dto

import "time"

const (
	StatusPlanned = "planned"
	DirectionNext = "next"
	DirectionPrev = "prev"
)

type Response struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartsAt    time.Time `json:"starts_at"`
	EndsAt      time.Time `json:"ends_at"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListResponse struct {
	Data       []Response `json:"data"`
	NextCursor *int64     `json:"next_cursor"`
	PrevCursor *int64     `json:"prev_cursor"`
	HasNext    bool       `json:"has_next"`
	HasPrev    bool       `json:"has_prev"`
}
