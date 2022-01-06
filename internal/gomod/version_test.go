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
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLatestTag(t *testing.T) {
	dir := t.TempDir()
	_, err := exec.Command("git", "clone", "-b", "v0.4.0", "https://github.com/CycloneDX/cyclonedx-go.git", dir).CombinedOutput()
	require.NoError(t, err)

	v, err := GetVersionFromTag(dir)
	require.NoError(t, err)

	sv := strings.SplitN(v, "-", 3)

	require.Equal(t, "v0.4.1", sv[0])
	require.Equal(t, "dc02c3afeacc", sv[2])
}
