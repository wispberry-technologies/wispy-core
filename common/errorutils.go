package common

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidationErrorsToMessage(err error) string {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, fieldErr := range validationErrs {
			messages = append(messages, fmt.Sprintf("Field '%s': %s", fieldErr.Field(), fieldErr.Error()))
		}
		return strings.Join(messages, ", ")
	}
	return err.Error()
}
