package pkg

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/pkg/gomod"
	"github.com/stretchr/testify/require"
)

func TestToComponent(t *testing.T) {
	m := gomod.Module{
		Path:    "modulePath",
		Version: "moduleVersion",
	}

	p := gomod.Package{
		ImportPath: "packagePath",
	}

	c, err := ToComponent(p, m)
	require.NoError(t, err)
	require.Equal(t, "packagePath", c.Name)
	require.Equal(t, "moduleVersion", c.Version)
	require.Equal(t, "pkg:golang/packagePath@moduleVersion?type=package", c.PackageURL)
}
