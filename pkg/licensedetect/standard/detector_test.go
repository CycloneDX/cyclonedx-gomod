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

package standard

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetector_Detect(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		licenses, err := NewDetector(zerolog.Nop()).Detect("path", "version", "../../../")
		require.NoError(t, err)
		require.Len(t, licenses, 1)
		assert.Equal(t, "Apache-2.0", licenses[0].ID)
	})

	t.Run("NoLicenseDetected", func(t *testing.T) {
		licenses, err := NewDetector(zerolog.Nop()).Detect("path", "version", t.TempDir())
		require.NoError(t, err)
		require.Empty(t, licenses)
	})
}
