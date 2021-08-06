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

package license

import (
	"os"
	"strings"
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		licenses, err := Resolve(gomod.Module{
			Dir: "../../",
		})
		require.NoError(t, err)
		require.Len(t, licenses, 1)
		assert.Equal(t, "Apache-2.0", licenses[0].ID)
	})

	t.Run("NoLicenseFound", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", strings.ReplaceAll(t.Name()+"_*", "/", "_"))
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		_, err = Resolve(gomod.Module{
			Dir: tmpDir,
		})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrNoLicenseFound)
	})
}
