package request

import "github.com/go-playground/validator/v10"

type CreateScript struct {
	Command string `json:"command" validate:"required,min=1"`
}

func (cs *CreateScript) Validate(valid *validator.Validate) error { return valid.Struct(cs) }
