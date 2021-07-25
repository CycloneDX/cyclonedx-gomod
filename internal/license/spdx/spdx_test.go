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

package spdx

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLicenseByID(t *testing.T) {
	require.NotEmpty(t, licenses)

	license := GetLicenseByID("MIT")
	require.NotNil(t, license)
	assert.Equal(t, "MIT", license.ID)
	assert.Equal(t, "https://spdx.org/licenses/MIT.html", license.Reference)

	license = GetLicenseByID("doesNotExist")
	assert.Nil(t, license)
}
