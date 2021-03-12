package gomod

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModule_ParseModuleGraph(t *testing.T) {
	rawGraph := `github.com/CycloneDX/cyclonedx-go github.com/bradleyjkemp/cupaloy/v2@v2.6.0
github.com/CycloneDX/cyclonedx-go github.com/stretchr/testify@v1.7.0
github.com/bradleyjkemp/cupaloy/v2@v2.6.0 github.com/davecgh/go-spew@v1.1.1
github.com/bradleyjkemp/cupaloy/v2@v2.6.0 github.com/pmezard/go-difflib@v1.0.0
github.com/bradleyjkemp/cupaloy/v2@v2.6.0 github.com/stretchr/objx@v0.1.1
github.com/bradleyjkemp/cupaloy/v2@v2.6.0 github.com/stretchr/testify@v1.6.1
github.com/stretchr/testify@v1.7.0 github.com/davecgh/go-spew@v1.1.0
github.com/stretchr/testify@v1.7.0 github.com/pmezard/go-difflib@v1.0.0
github.com/stretchr/testify@v1.7.0 github.com/stretchr/objx@v0.1.0
github.com/stretchr/testify@v1.7.0 gopkg.in/yaml.v3@v3.0.0-20200313102051-9f266ea9e77c
github.com/stretchr/testify@v1.6.1 github.com/davecgh/go-spew@v1.1.0
github.com/stretchr/testify@v1.6.1 github.com/pmezard/go-difflib@v1.0.0
github.com/stretchr/testify@v1.6.1 github.com/stretchr/objx@v0.1.0
github.com/stretchr/testify@v1.6.1 gopkg.in/yaml.v3@v3.0.0-20200313102051-9f266ea9e77c
gopkg.in/yaml.v3@v3.0.0-20200313102051-9f266ea9e77c gopkg.in/check.v1@v0.0.0-20161208181325-20d25e280405
`
	module := Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	}

	graph, err := module.parseModuleGraph(strings.NewReader(rawGraph))
	require.NoError(t, err)

	directDependencies, ok := graph["github.com/CycloneDX/cyclonedx-go@v0.1.0"]
	require.True(t, ok)
	assert.Contains(t, directDependencies, "github.com/bradleyjkemp/cupaloy/v2@v2.6.0")
	assert.Contains(t, directDependencies, "github.com/stretchr/testify@v1.7.0")

	transitiveDependencies, ok := graph["github.com/bradleyjkemp/cupaloy/v2@v2.6.0"]
	require.True(t, ok)
	assert.Contains(t, transitiveDependencies, "github.com/davecgh/go-spew@v1.1.1")
	assert.Contains(t, transitiveDependencies, "github.com/pmezard/go-difflib@v1.0.0")
	assert.Contains(t, transitiveDependencies, "github.com/stretchr/objx@v0.1.1")
	assert.Contains(t, transitiveDependencies, "github.com/stretchr/testify@v1.6.1")
}

func TestModule_PackageURL(t *testing.T) {
	module := Module{
		Path:    "github.com/CycloneDX/cyclonedx-go",
		Version: "v0.1.0",
	}

	assert.Equal(t, "pkg:golang/github.com/CycloneDX/cyclonedx-go@v0.1.0", module.PackageURL())
}

func TestParseModules(t *testing.T) {
	modulesJSON := `{
        "Path": "github.com/CycloneDX/cyclonedx-go",
        "Main": true,
        "Dir": "/path/to/cyclonedx-go",
        "GoMod": "/path/to/cyclonedx-go/go.mod",
        "GoVersion": "1.14"
}
{
        "Path": "github.com/davecgh/go-spew",
        "Version": "v1.1.1",
        "Time": "2018-02-21T23:26:28Z",
        "Indirect": true,
        "Dir": "/path/to/go/pkg/mod/github.com/davecgh/go-spew@v1.1.1",
        "GoMod": "/path/to/go/pkg/mod/cache/download/github.com/davecgh/go-spew/@v/v1.1.1.mod"
}`

	modules, err := parseModules(strings.NewReader(modulesJSON))
	require.NoError(t, err)

	assert.Len(t, modules, 2)
	assert.Equal(t, "github.com/CycloneDX/cyclonedx-go", modules[0].Path)
	assert.True(t, modules[0].Main)
	assert.Equal(t, "github.com/davecgh/go-spew", modules[1].Path)
	assert.Equal(t, "v1.1.1", modules[1].Version)
	assert.False(t, modules[1].Main)
}
