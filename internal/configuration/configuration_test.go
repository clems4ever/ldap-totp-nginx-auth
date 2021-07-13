package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/utils"
)

func TestShouldHaveSameChecksumForBothTemplates(t *testing.T) {
	sumRoot, err := utils.HashSHA256FromPath("../../config.template.yml")
	assert.NoError(t, err)

	sumInternal, err := utils.HashSHA256FromPath("./config.template.yml")
	assert.NoError(t, err)

	assert.Equal(t, sumRoot, sumInternal)
}
