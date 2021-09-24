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

package util

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/stretchr/testify/require"
)

func TestSetSerialNumber(t *testing.T) {
	t.Run("NoSerialNumber", func(t *testing.T) {
		require.NoError(t, SetSerialNumber(nil, options.SBOMOptions{
			SerialNumber:   "",
			NoSerialNumber: true,
		}))
	})

	t.Run("DefaultRandomSerialNumber", func(t *testing.T) {
		bom := new(cyclonedx.BOM)

		require.NoError(t, SetSerialNumber(bom, options.SBOMOptions{
			SerialNumber:   "",
			NoSerialNumber: false,
		}))
		require.Regexp(t, `^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, bom.SerialNumber)
	})

	t.Run("CustomSerialNumber", func(t *testing.T) {
		bom := new(cyclonedx.BOM)

		require.NoError(t, SetSerialNumber(bom, options.SBOMOptions{
			SerialNumber:   "00000000-0000-0000-0000-000000000000",
			NoSerialNumber: false,
		}))
		require.Equal(t, "urn:uuid:00000000-0000-0000-0000-000000000000", bom.SerialNumber)
	})

	t.Run("InvalidCustomSerialNumber", func(t *testing.T) {
		require.Error(t, SetSerialNumber(new(cyclonedx.BOM), options.SBOMOptions{
			SerialNumber:   "invalid",
			NoSerialNumber: false,
		}))
	})
}
