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

	"github.com/stretchr/testify/require"
)

func TestParseModulesFromBinary(t *testing.T) {
	cmdOutput := `minikube: go1.16.4
path    k8s.io/minikube/cmd/minikube
mod     k8s.io/minikube (devel) 
dep     cloud.google.com/go     v0.84.0 h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=
dep     github.com/briandowns/spinner   v1.11.1
=>      github.com/alonyb/spinner       v1.12.7 h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=
`

	goVersion, modules, hashes := parseModulesFromBinary("minikube", strings.NewReader(cmdOutput))
	require.Equal(t, "go1.16.4", goVersion)
	require.Len(t, modules, 3)
	require.Len(t, hashes, 2)

	// Main module
	require.Equal(t, "k8s.io/minikube", modules[0].Path)
	require.Equal(t, "(devel)", modules[0].Version)
	require.True(t, modules[0].Main)
	require.NotContains(t, hashes, modules[0].Coordinates())

	// Module w/o replacement
	require.Equal(t, "cloud.google.com/go", modules[1].Path)
	require.Equal(t, "v0.84.0", modules[1].Version)
	require.Contains(t, hashes, modules[1].Coordinates())
	require.Equal(t, "h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=", hashes["cloud.google.com/go@v0.84.0"])

	// Module with replacement
	require.Equal(t, "github.com/briandowns/spinner", modules[2].Path)
	require.Equal(t, "v1.11.1", modules[2].Version)
	require.NotContains(t, hashes, modules[2].Coordinates())
	require.NotNil(t, modules[2].Replace)

	// Replacement
	require.Equal(t, "github.com/alonyb/spinner", modules[2].Replace.Path)
	require.Equal(t, "v1.12.7", modules[2].Replace.Version)
	require.Contains(t, hashes, modules[2].Replace.Coordinates())
	require.Equal(t, "h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=", hashes["github.com/alonyb/spinner@v1.12.7"])
}
