package validate

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strings"
)

var validate = validator.New()

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

type ErrorResponse struct {
	FailedField string
	Tag         string
	Value       string
}

func Req(data any) []*ErrorResponse {
	var errors []*ErrorResponse

	err := validate.Struct(data)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.Namespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
