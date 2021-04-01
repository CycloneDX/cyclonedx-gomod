package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileExists(t *testing.T) {
	assert.False(t, FileExists("doesNotExist"))

	tmpFile, err := os.CreateTemp("", "TestFileExists_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	require.True(t, FileExists(tmpFile.Name()))
}

func TestIsGoModule(t *testing.T) {
	assert.True(t, IsGoModule("../../"))

	tmpDir, err := os.MkdirTemp("", "TestIsGoModule_*")
	require.NoError(t, err)
	defer os.Remove(tmpDir)
	require.False(t, IsGoModule(tmpDir))
}

func TestStartsWith(t *testing.T) {
	assert.True(t, StartsWith("startsWithSomething", "startsWithSomething"))
	assert.True(t, StartsWith("startsWithSomething", "startsWith"))
	assert.False(t, StartsWith("startsWithSomething", " startsWith"))
}
