package logging

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func TestShouldWriteLogsToFile(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "logs-dir")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	path := fmt.Sprintf("%s/authelia.log", dir)
	err = InitializeLogger(schema.LogConfiguration{Format: "text", FilePath: path, KeepStdout: false}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "level=info msg=\"This is a test\"\n")
}

func TestShouldWriteLogsToFileAndStdout(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "logs-dir")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	path := fmt.Sprintf("%s/authelia.log", dir)
	err = InitializeLogger(schema.LogConfiguration{Format: "text", FilePath: path, KeepStdout: true}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "level=info msg=\"This is a test\"\n")
}

func TestShouldFormatLogsAsJSON(t *testing.T) {
	dir, err := ioutil.TempDir("/tmp", "logs-dir")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)

	path := fmt.Sprintf("%s/authelia.log", dir)
	err = InitializeLogger(schema.LogConfiguration{Format: "json", FilePath: path, KeepStdout: false}, false)
	require.NoError(t, err)

	Logger().Info("This is a test")

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)

	assert.Contains(t, string(b), "{\"level\":\"info\",\"msg\":\"This is a test\",")
}
