package verve

import (
	validation "github.com/go-ozzo/ozzo-validation"
)

func validateId(id string) error {
	return validation.Validate(id, validation.Required)
}
