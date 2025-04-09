package verve

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSuccessfulValidation(t *testing.T) {
	err := validateId("id")
	assert.NoError(t, err)
}

func TestValidationFailureWithMissingID(t *testing.T) {
	err := validateId("")

	expectedErr := "cannot be blank"
	assert.EqualError(t, err, expectedErr)
}
