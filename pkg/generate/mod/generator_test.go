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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGenerator(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		g, err := NewGenerator("")
		require.NoError(t, err)
		require.NotNil(t, g)
	})

	t.Run("OptionError", func(t *testing.T) {
		failOption := func(g *generator) error {
			return errors.New("test")
		}

		g, err := NewGenerator("", failOption)
		require.Nil(t, g)
		require.Error(t, err)
		require.Equal(t, "test", err.Error())
	})
}
