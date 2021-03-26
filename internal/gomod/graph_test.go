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

func TestGetEffectiveModuleGraph(t *testing.T) {
	moduleGraph := map[string][]string{
		"github.com/acme-inc/acme-app@v1.0.0": {
			"github.com/acme-inc/acme-lib@v1.1.1",
			"golang.org/x/crypto@v0.0.0-20200622213623-75b288015ac9",
		},
		"github.com/acme-inc/acme-lib@v1.1.1":                    {},
		"golang.org/x/crypto@v0.0.0-20200622213623-75b288015ac9": {},
	}

	modules := []Module{
		{Path: "github.com/acme-inc/acme-app", Version: "v1.0.0"},
		{Path: "github.com/acme-inc/acme-lib", Version: "v1.1.3"},
		{
			Path:    "golang.org/x/crypto",
			Version: "v0.0.0-20200622213623-75b288015ac9",
			Replace: &Module{
				Path:    "github.com/acme-inc/acme-crypto",
				Version: "v0.0.0-20200622213623-75b288015ac9",
			},
		},
	}

	effectiveGraph, err := GetEffectiveModuleGraph(moduleGraph, modules)
	require.NoError(t, err)

	assert.Equal(t, "github.com/acme-inc/acme-lib@v1.1.3", effectiveGraph["github.com/acme-inc/acme-app@v1.0.0"][0])
	assert.Equal(t, "github.com/acme-inc/acme-crypto@v0.0.0-20200622213623-75b288015ac9", effectiveGraph["github.com/acme-inc/acme-app@v1.0.0"][1])
	assert.Empty(t, effectiveGraph["github.com/acme-inc/acme-lib@v1.1.1"])
	assert.Empty(t, effectiveGraph["github.com/acme-inc/acme-crypto@v0.0.0-20200622213623-75b288015ac9"])
}
