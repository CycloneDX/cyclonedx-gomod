package gocmd

import (
	"bytes"
	"encoding/json"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVersion(t *testing.T) {
	version, err := GetVersion()
	require.NoError(t, err)
	require.Equal(t, runtime.Version(), version)
}

func TestGetModuleList(t *testing.T) {
	buf := new(bytes.Buffer)
	err := GetModuleList("../../", buf)
	require.NoError(t, err)

	mod := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(buf).Decode(&mod))

	// Smoke test - is this really the module list?
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-gomod", mod["Path"])
	assert.Equal(t, true, mod["Main"])
}

func TestGetModuleGraph(t *testing.T) {
	buf := new(bytes.Buffer)
	err := GetModuleGraph("../../", buf)
	require.NoError(t, err)

	assert.Equal(t, 0, strings.Index(buf.String(), "github.com/CycloneDX/cyclonedx-gomod"))
}
