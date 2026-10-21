package field_dto

type Validation struct {
	Required  bool     `json:"required"`
	MinLength *int     `json:"min_length,omitempty" validate:"omitempty,gte=0"`
	MaxLength *int     `json:"max_length,omitempty" validate:"omitempty,gte=0"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
}

type Create struct {
	PageID     int64      `json:"page_id" validate:"required,gt=0"`
	Sort       int        `json:"sort" validate:"gte=1,lte=2147483647"`
	Type       string     `json:"type" validate:"required,oneof=text textarea email tel number date select radio checkbox file"`
	Options    []string   `json:"options" validate:"max=1000,dive,required,max=200"`
	Name       string     `json:"name" validate:"required,max=200"`
	Validation Validation `json:"validation"`
	IsActive   *bool      `json:"is_active" validate:"required"`
}

type Update Create

type Filter struct {
	PageID   int64
	Name     string
	Type     string
	IsActive *bool
}
