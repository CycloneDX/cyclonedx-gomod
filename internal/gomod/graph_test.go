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

package gomod

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindModule(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		modules := make([]Module, 0)
		require.Nil(t, findModule(modules, "path@version", true))
	})

	t.Run("Strict", func(t *testing.T) {
		modules := []Module{
			{
				Path:    "path",
				Version: "version1",
			},
			{
				Path:    "path",
				Version: "version2",
			},
		}

		require.Nil(t, findModule(modules, "path@version0", true))
		require.Nil(t, findModule(modules, "otherpath@version1", true))

		module := findModule(modules, "path@version2", true)
		require.NotNil(t, module)
		require.Equal(t, "path@version2", module.Coordinates())
	})

	t.Run("NonStrict", func(t *testing.T) {
		modules := []Module{
			{
				Path:    "path",
				Version: "version1",
			},
			{
				Path:    "path",
				Version: "version2",
			},
		}

		module := findModule(modules, "path@version0", false)
		require.NotNil(t, module)
		require.Equal(t, "path@version1", module.Coordinates())

		require.Nil(t, findModule(modules, "otherpath@version1", false))

		require.Equal(t, module, findModule(modules, "path@version2", false))
	})
}

func TestSortDependencies(t *testing.T) {
	modules := []*Module{
		{
			Path:    "path",
			Version: "v1.3.2",
		},
		{
			Path:    "path",
			Version: "v1.2.3",
		},
		{
			Path:    "path/v2",
			Version: "v2.0.0",
			Main:    true,
		},
	}

	sortDependencies(modules)

	require.Equal(t, "v1.2.3", modules[0].Version)
	require.Equal(t, "v1.3.2", modules[1].Version)
	require.Equal(t, "v2.0.0", modules[2].Version) // main
}
