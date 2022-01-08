package pkg

import (
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
)

func TestToComponent(t *testing.T) {
	m := gomod.Module{
		Path:    "modulePath",
		Version: "moduleVersion",
	}

	p := gomod.Package{
		ImportPath: "packagePath",
	}

	c, err := ToComponent(zerolog.New(io.Discard), p, m)
	require.NoError(t, err)
	require.Equal(t, "packagePath", c.Name)
	require.Equal(t, "moduleVersion", c.Version)
	require.Equal(t, "pkg:golang/packagePath@moduleVersion?type=package", c.PackageURL)
}
