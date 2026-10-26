package category_dto

import (
	"encoding/json"
	"time"
)

type Response struct {
	ID           int64     `json:"id"`
	TournamentID int64     `json:"tournament_id"`
	ParentID     *int64    `json:"parent_id"`
	Name         string    `json:"name"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ListResponse struct {
	Data       []Response `json:"data"`
	NextCursor *int64     `json:"next_cursor"`
	PrevCursor *int64     `json:"prev_cursor"`
	HasNext    bool       `json:"has_next"`
	HasPrev    bool       `json:"has_prev"`
}

type WithChildrenResponse struct {
	Response
	Children json.RawMessage `json:"children"`
}

type WithChildrenListResponse struct {
	Data []*WithChildrenResponse `json:"data"`
}
