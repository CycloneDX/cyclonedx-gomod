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

package bin

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/require"
)

func TestWithModuleHashes(t *testing.T) {
	t.Run("HashAvailable", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
		}

		hashes := map[string]string{
			"path@version": "h1:qJYtXnJRWmpe7m/3XlyhrsLrEURqHRM2kxzoxXqyUDs=",
		}

		component := new(cdx.Component)

		err := withModuleHashes(hashes)(module, component)
		require.NoError(t, err)
		require.NotNil(t, component.Hashes)
		require.Equal(t, cdx.HashAlgoSHA256, (*component.Hashes)[0].Algorithm)
		require.Equal(t, "a8962d5e72515a6a5eee6ff75e5ca1aec2eb11446a1d1336931ce8c57ab2503b", (*component.Hashes)[0].Value)
	})

	t.Run("HashNotAvailable", func(t *testing.T) {
		module := gomod.Module{
			Path:    "path",
			Version: "version",
		}

		hashes := make(map[string]string)

		component := new(cdx.Component)

		err := withModuleHashes(hashes)(module, component)
		require.NoError(t, err)
		require.Nil(t, component.Hashes)
	})
}
