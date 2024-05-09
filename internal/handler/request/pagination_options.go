package request

import (
	"math"

	"github.com/go-playground/validator/v10"
)

type PaginationOptions struct {
	Offset int `json:"offset" validate:"min=0"`
	Limit  int `json:"limit" validate:"required,min=0"`
}

func (po *PaginationOptions) Validate(valid *validator.Validate) error {
	return valid.Struct(po)
}

func GetUnlimitedPaginationOptions() PaginationOptions {
	return PaginationOptions{
		Offset: 0,
		Limit:  math.MaxInt64,
	}
}
