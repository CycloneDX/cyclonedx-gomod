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
	"io"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestLoadStdlibModule(t *testing.T) {
	module, err := LoadStdlibModule(zerolog.New(io.Discard))
	require.NoError(t, err)
	require.Equal(t, "std", module.Path)
	require.Regexp(t, `^go\d\.`, module.Version)

	require.Nil(t, module.Replace)
	require.NotEmpty(t, module.Dir)
	require.False(t, module.Indirect)
	require.Empty(t, module.Dependencies)
	require.True(t, module.Local)
	require.False(t, module.Main)
	require.Empty(t, module.Packages)
	require.False(t, module.TestOnly)
	require.False(t, module.Vendored)
}
