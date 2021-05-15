package license

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	// Success with single license
	licenses, err := Resolve(gomod.Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	})
	require.NoError(t, err)
	require.Len(t, licenses, 1)
	assert.Equal(t, "Apache-2.0", licenses[0].ID)

	// Success with multiple licenses
	licenses, err = Resolve(gomod.Module{
		Path:    "gopkg.in/yaml.v3",
		Version: "v3.0.0-20200313102051-9f266ea9e77c",
	})
	require.NoError(t, err)
	require.Len(t, licenses, 2)
	assert.Equal(t, "Apache-2.0", licenses[0].ID)
	assert.Equal(t, "MIT", licenses[1].ID)

	// Module not found
	_, err = Resolve(gomod.Module{
		Path:    "github.com/CycloneDX/doesnotexist",
		Version: "v1.0.0",
	})
	assert.ErrorIs(t, err, ErrModuleNotFound)
}
