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
// Copyright (c) Niklas Düster. All Rights Reserved.

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
	assert.NotEmpty(t, licenses[0].URL)

	// Success with multiple licenses
	licenses, err = Resolve(gomod.Module{
		Path:    "github.com/BurntSushi/xgb",
		Version: "v0.0.0-20160522181843-27f122750802",
	})
	require.NoError(t, err)
	require.Len(t, licenses, 2)
	assert.Equal(t, "BSD-3-Clause", licenses[0].ID)
	assert.Empty(t, licenses[0].Name)
	assert.NotEmpty(t, licenses[0].URL)
	assert.Empty(t, licenses[1].ID)
	assert.Equal(t, "GooglePatentClause", licenses[1].Name)
	assert.Empty(t, licenses[1].URL)

	// Module not found
	_, err = Resolve(gomod.Module{
		Path:    "github.com/CycloneDX/doesnotexist",
		Version: "v1.0.0",
	})
	assert.ErrorIs(t, err, ErrModuleNotFound)
}
