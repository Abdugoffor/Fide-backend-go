package category_dto

type Create struct {
	TournamentID int64  `json:"tournament_id" validate:"required,gt=0"`
	ParentID     *int64 `json:"parent_id" validate:"omitempty,gt=0"`
	Name         string `json:"name" validate:"required,max=200"`
	IsActive     *bool  `json:"is_active" validate:"required"`
}

type Update Create

type Filter struct {
	Cursor       int64
	Limit        int
	TournamentID int64
	ParentID     *int64
	Name         string
	IsActive     *bool
	Direction    string
}
