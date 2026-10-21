package page_dto

type Create struct {
	TournamentID int64  `json:"tournament_id" validate:"required,gt=0"`
	CategoryID   int64  `json:"category_id" validate:"required,gt=0"`
	Name         string `json:"name" validate:"required,max=200"`
	Sort         int    `json:"sort" validate:"gte=3,lte=2147483647"`
	IsActive     *bool  `json:"is_active" validate:"required"`
}

type Update Create

type Filter struct {
	TournamentID int64
	CategoryID   int64
	Name         string
	IsActive     *bool
}
