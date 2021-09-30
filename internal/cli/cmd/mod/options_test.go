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

package mod

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/stretchr/testify/require"
)

func TestModOptions_Validate(t *testing.T) {
	t.Run("InvalidComponentType", func(t *testing.T) {
		var modOptions Options
		modOptions.ComponentType = "foobar"

		err := modOptions.Validate()
		require.Error(t, err)

		var validationError *options.ValidationError
		require.ErrorAs(t, err, &validationError)

		require.Len(t, validationError.Errors, 1)
		require.Contains(t, validationError.Errors[0].Error(), "component type: \"foobar\" is invalid")
	})
}
