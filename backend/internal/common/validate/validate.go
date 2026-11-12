package validate

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var tagMessages = map[string]string{
	"required": "is required",
	"min":      "is too short",
	"max":      "is too long",
	"email":    "is not a valid email",
	"alphanum": "must contain only letters and numbers",
	"eqfield":  "does not match",
	"datetime": "is not a valid date",
	"oneof":    "is invalid",
	"e164":     "is not a valid phone number",
	"len":      "has the wrong length",
	"numeric":  "must contain only numbers",
}

func ParseErrors(err error) map[string]string {
	errs := make(map[string]string)

	ve, ok := err.(validator.ValidationErrors)
	if !ok {
		errs["message"] = err.Error()
		return errs
	}

	for _, fe := range ve {
		field := toLowerCamel(fe.Field())
		msg, ok := tagMessages[fe.Tag()]
		if !ok {
			msg = fmt.Sprintf("failed validation on '%s'", fe.Tag())
		}
		errs[field] = msg
	}
	return errs
}

func toLowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]+32) + s[1:]
}
