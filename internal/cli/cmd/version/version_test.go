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

package version

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecVersionCmd(t *testing.T) {
	t.Run("Plain", func(t *testing.T) {
		buf := new(bytes.Buffer)

		err := execVersionCmd(buf, false)
		require.NoError(t, err)

		lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
		require.Len(t, lines, 4)
		require.Equal(t, "Version:\tv0.0.0-unknown", lines[0]) // Actual version is not set during tests
		require.Regexp(t, `^GoVersion:\s+go.+`, lines[1])
		require.Regexp(t, `^OS:\s+[a-z]+`, lines[2])
		require.Regexp(t, `^Arch:\s+[a-z]+`, lines[3])
	})

	t.Run("JSON", func(t *testing.T) {
		buf := new(bytes.Buffer)

		err := execVersionCmd(buf, true)
		require.NoError(t, err)

		var output map[string]any
		err = json.NewDecoder(buf).Decode(&output)
		require.NoError(t, err)

		require.Len(t, output, 4)
		require.Contains(t, output, "Version")
		require.Equal(t, "v0.0.0-unknown", output["Version"])
		require.Contains(t, output, "GoVersion")
		require.Regexp(t, `go.+`, output["GoVersion"])
		require.Contains(t, output, "OS")
		require.Contains(t, output, "Arch")
	})
}
