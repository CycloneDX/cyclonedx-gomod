// This file is part of CycloneDX GoMod
//
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
// Copyright (c) OWASP Foundation. All Rights Reserved.

package module

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithLicenses(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		module := gomod.Module{
			Dir: "../../../",
		}
		component := cdx.Component{}

		err := WithLicenses()(module, &component)
		require.NoError(t, err)
		require.NotNil(t, component.Evidence)
		require.NotNil(t, component.Evidence.Licenses)
		require.Len(t, *component.Evidence.Licenses, 1)
	})

	t.Run("Not Found", func(t *testing.T) {
		module := gomod.Module{
			Dir: ".",
		}
		component := cdx.Component{}

		err := WithLicenses()(module, &component)
		require.NoError(t, err)
		require.Nil(t, component.Evidence)
	})

	t.Run("Other Error", func(t *testing.T) {
		module := gomod.Module{
			Dir: "./doesNotExist",
		}
		component := cdx.Component{}

		err := WithLicenses()(module, &component)
		require.Error(t, err)
		require.Nil(t, component.Evidence)
	})
}

func TestWithComponentType(t *testing.T) {
	module := gomod.Module{}
	component := cdx.Component{}

	err := WithComponentType(cdx.ComponentTypeContainer)(module, &component)
	require.NoError(t, err)
	require.Equal(t, cdx.ComponentTypeContainer, component.Type)
}

func TestWithScope(t *testing.T) {
	module := gomod.Module{}
	component := cdx.Component{}

	err := WithScope(cdx.ScopeExcluded)(module, &component)
	require.NoError(t, err)
	require.Equal(t, cdx.ScopeExcluded, component.Scope)
}

func TestWithTestScope(t *testing.T) {
	t.Run("TestOnly", func(t *testing.T) {
		module := gomod.Module{
			TestOnly: true,
		}
		component := cdx.Component{}

		err := WithTestScope(cdx.ScopeExcluded)(module, &component)
		require.NoError(t, err)
		require.Equal(t, cdx.ScopeExcluded, component.Scope)
	})

	t.Run("Not TestOnly", func(t *testing.T) {
		module := gomod.Module{
			TestOnly: false,
		}
		component := cdx.Component{}

		err := WithTestScope(cdx.ScopeExcluded)(module, &component)
		require.NoError(t, err)
		require.Equal(t, cdx.Scope(""), component.Scope)
	})
}

func TestToComponent(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
		}

		component, err := ToComponent(module)
		require.NoError(t, err)
		require.NotNil(t, component)

		require.Equal(t, "pkg:golang/path@version", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "path", component.Name)
		require.Equal(t, "version", component.Version)
		require.Equal(t, "pkg:golang/path@version", component.PackageURL)
		require.Equal(t, cdx.ScopeRequired, component.Scope)
	})

	t.Run("With TestOnly", func(t *testing.T) {
		module := gomod.Module{
			Path:     "path",
			Version:  "version",
			TestOnly: true,
		}

		component, err := ToComponent(module)
		require.NoError(t, err)
		require.NotNil(t, component)

		require.Equal(t, "pkg:golang/path@version", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "path", component.Name)
		require.Equal(t, "version", component.Version)
		require.Equal(t, "pkg:golang/path@version", component.PackageURL)
		require.Equal(t, cdx.ScopeOptional, component.Scope)
	})

	t.Run("With Replace", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
			Replace: &gomod.Module{
				Path:    "pathReplace",
				Version: "versionReplace",
			},
		}

		component, err := ToComponent(module)
		require.NoError(t, err)
		require.NotNil(t, component)

		require.Equal(t, "pkg:golang/pathReplace@versionReplace", component.BOMRef)
		require.Equal(t, cdx.ComponentTypeLibrary, component.Type)
		require.Equal(t, "pathReplace", component.Name)
		require.Equal(t, "versionReplace", component.Version)
		require.Equal(t, "pkg:golang/pathReplace@versionReplace", component.PackageURL)
		require.Equal(t, cdx.ScopeRequired, component.Scope)
	})
}

func TestResolveVCSURL(t *testing.T) {
	t.Run("GitHub", func(t *testing.T) {
		require.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go"))
	})

	t.Run("GitHub with major version", func(t *testing.T) {
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go/v2"))
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", resolveVCSURL("github.com/CycloneDX/cyclonedx-go/v222"))
	})

	t.Run("gopkg.in variant 1", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-playground/assert", resolveVCSURL("gopkg.in/go-playground/assert.v1"))
	})

	t.Run("gopkg.in variant 2", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-check/check", resolveVCSURL("gopkg.in/check.v1"))
	})
}
