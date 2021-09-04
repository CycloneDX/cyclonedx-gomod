package app

import (
	"os"
	"runtime"
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateBuildProperties(t *testing.T) {
	origGoflags := os.Getenv("GOFLAGS")
	os.Setenv("GOFLAGS", "-tags=foo,bar")

	if origGoflags != "" {
		defer func() {
			os.Setenv("GOFLAGS", origGoflags)
		}()
	}

	properties, err := createBuildProperties()
	require.NoError(t, err)
	require.Len(t, properties, 6)

	expectedCgoEnabled := "1" // Cgo is enabled per default
	if cgo := os.Getenv("CGO_ENABLED"); cgo != "" {
		expectedCgoEnabled = cgo
	}

	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:CGO_ENABLED", Value: expectedCgoEnabled})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOARCH", Value: runtime.GOARCH})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOOS", Value: runtime.GOOS})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOVERSION", Value: runtime.Version()})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "foo"})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "bar"})
}
