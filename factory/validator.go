package factory

import (
	"fmt"
	"strings"

	"github.com/hublabs/common/api"
	"gopkg.in/go-playground/validator.v9"
)

type Validator struct {
	validator *validator.Validate
}

func (cv *Validator) Validate(i interface{}) error {
	err := cv.validator.Struct(i)
	if err == nil {
		return nil
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		msg := make([]string, 0)
		for _, err := range errs {
			msg = append(msg, fmt.Sprintf("%v condition: %v ,value: %v", err.Field(), err.ActualTag(), err.Value()))
		}
		return NewError(api.ErrorInvalidFields, strings.Join(msg, ","))
	}
	return err
}

func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}
