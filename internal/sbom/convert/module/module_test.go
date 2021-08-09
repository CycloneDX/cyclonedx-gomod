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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
