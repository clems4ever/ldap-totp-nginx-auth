package configuration

import (
	"os"
	"runtime"
	"testing"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestShouldErrorPermissionsOnLocalFS(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	_ = os.Mkdir("/tmp/noperms/", 0000)
	created, err := GenerateIfNotFound("/tmp/noperms/configuration.yml")
	require.Equal(t, created, false)
	require.EqualError(t, err, "Unable to generate /tmp/noperms/configuration.yml: open /tmp/noperms/configuration.yml: permission denied")
}

func TestShouldGenerateConfigIfNotFound(t *testing.T) {
	if runtime.GOOS == windows {
		t.Skip("skipping test due to being on windows")
	}

	dir := "/tmp/authelia" + utils.RandomString(10, authentication.HashingPossibleSaltCharacters) + "/"
	err := os.MkdirAll(dir, 0700)
	require.NoError(t, err)

	created, err := GenerateIfNotFound(dir + "non-existing-config.yml")
	require.Equal(t, created, true)
	require.Equal(t, err, nil)

	stat, err := os.Stat(dir + "non-existing-config.yml")
	require.Equal(t, err, nil)
	require.EqualValues(t, stat.Size(), len(configTemplate))
}
