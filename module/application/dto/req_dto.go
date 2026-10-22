package application_dto

type Create struct {
	TournamentID int64  `json:"tournament_id" validate:"required,gt=0"`
	CategoryID   int64  `json:"category_id" validate:"required,gt=0"`
	Name         string `json:"name" validate:"required,max=200"`
	Email        string `json:"email" validate:"required,email,max=254"`
	BirthDate    string `json:"birth_date" validate:"required,datetime=2006-01-02"`
}

type VerifyEmail struct {
	TournamentID  int64  `json:"tournament_id" validate:"required,gt=0"`
	Email         string `json:"email" validate:"required,email,max=254"`
	ApplicationID int64  `json:"application_id" validate:"required,gt=0"`
	Code          string `json:"code" validate:"required,len=6,numeric"`
}
