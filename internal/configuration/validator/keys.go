package validator

import (
	"errors"
	"fmt"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateKeys determines if all provided keys are valid.
func ValidateKeys(keys []string, prefix string, validator *schema.StructValidator) {
	var errStrings []string

	for _, key := range keys {
		expectedKey := reKeyReplacer.ReplaceAllString(key, "[]")

		if utils.IsStringInSlice(expectedKey, ValidKeys) {
			continue
		}

		if newKey, ok := replacedKeys[expectedKey]; ok {
			validator.Push(fmt.Errorf(errFmtReplacedConfigurationKey, key, newKey))
			continue
		}

		if err, ok := specificErrorKeys[expectedKey]; ok {
			if !utils.IsStringInSlice(err, errStrings) {
				errStrings = append(errStrings, err)
			}
		} else {
			if strings.HasPrefix(key, prefix) {
				validator.Push(fmt.Errorf("configuration environment variable not expected: %s", key))
			} else {
				validator.Push(fmt.Errorf("configuration key not expected: %s", key))
			}
		}
	}

	for _, err := range errStrings {
		validator.Push(errors.New(err))
	}
}
