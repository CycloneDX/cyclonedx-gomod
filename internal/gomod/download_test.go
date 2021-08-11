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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDownload(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", strings.ReplaceAll(t.Name(), "/", "_")+"_*")
		require.NoError(t, err)

		modcache := os.Getenv("GOMODCACHE")
		defer func() {
			os.RemoveAll(tmpDir)

			if modcache != "" {
				os.Setenv("GOMODCACHE", modcache)
			} else {
				os.Unsetenv("GOMODCACHE")
			}
		}()

		os.Setenv("GOMODCACHE", tmpDir)

		downloads, err := Download([]Module{
			{
				Path:    "github.com/CycloneDX/cyclonedx-go",
				Version: "v0.4.0",
			},
		})
		require.NoError(t, err)
		require.Len(t, downloads, 1)
		require.Empty(t, downloads[0].Error)
		require.Equal(t, filepath.Join(tmpDir, "github.com/!cyclone!d!x/cyclonedx-go@v0.4.0"), downloads[0].Dir)
		require.Equal(t, "h1:Wz4QZ9B4RXGWIWTypVLEOVJgOdFfy5mcS5PGNzUkZxU=", downloads[0].Sum)
	})

	t.Run("Error", func(t *testing.T) {
		downloads, err := Download([]Module{
			{
				Path:    "github.com/CycloneDX/cyclonedx-go-doesnotexist",
				Version: "",
			},
		})
		require.NoError(t, err)
		require.Len(t, downloads, 1)
		require.NotEmpty(t, downloads[0].Error)
	})
}
