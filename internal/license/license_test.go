package license

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	// Success
	license, err := Resolve(gomod.Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	})
	require.NoError(t, err)
	assert.Equal(t, "Apache-2.0", license)

	// Module not found
	_, err = Resolve(gomod.Module{
		Path:    "github.com/CycloneDX/doesnotexist",
		Version: "v1.0.0",
	})
	assert.ErrorIs(t, err, ErrModuleNotFound)
}
