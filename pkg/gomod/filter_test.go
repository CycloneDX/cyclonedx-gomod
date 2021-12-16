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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModWhy(t *testing.T) {
	modWhyOutput := `
# github.com/stretchr/testify
github.com/CycloneDX/cyclonedx-gomod
github.com/CycloneDX/cyclonedx-gomod.test
github.com/stretchr/testify/assert

# github.com/CycloneDX/cyclonedx-go
(main module does not need module github.com/CycloneDX/cyclonedx-go)

# bazil.org/fuse
(main module does not need to vendor module bazil.org/fuse)
`

	modulePkgs := parseModWhy(strings.NewReader(modWhyOutput))
	require.Len(t, modulePkgs, 3)

	assert.Len(t, modulePkgs["github.com/stretchr/testify"], 3)
	assert.Len(t, modulePkgs["github.com/CycloneDX/cyclonedx-go"], 0)
	assert.Len(t, modulePkgs["bazil.org/fuse"], 0)
}
