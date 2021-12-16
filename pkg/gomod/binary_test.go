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

func TestParseBuildInfo(t *testing.T) {
	t.Run("Go<1.18", func(t *testing.T) {
		out := `minikube: go1.16.4
path    k8s.io/minikube/cmd/minikube
mod     k8s.io/minikube (devel) 
dep     cloud.google.com/go     v0.84.0 h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=
dep     github.com/briandowns/spinner   v1.11.1
=>      github.com/alonyb/spinner       v1.12.7 h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=
`

		bi, err := parseBuildInfo("minikube", strings.NewReader(out))
		require.NoError(t, err)
		require.Equal(t, "go1.16.4", bi.GoVersion)
		require.Equal(t, "k8s.io/minikube/cmd/minikube", bi.Path)
		require.NotNil(t, bi.Main)
		require.Len(t, bi.Deps, 2)
		require.Nil(t, bi.Settings)

		// Main module
		require.Equal(t, "k8s.io/minikube", bi.Main.Path)
		require.Equal(t, "(devel)", bi.Main.Version)
		require.True(t, bi.Main.Main)
		require.Empty(t, bi.Main.Sum)

		// Module w/o replacement
		require.Equal(t, "cloud.google.com/go", bi.Deps[0].Path)
		require.Equal(t, "v0.84.0", bi.Deps[0].Version)
		require.Equal(t, "h1:hVhK90DwCdOAYGME/FJd9vNIZye9HBR6Yy3fu4js3N8=", bi.Deps[0].Sum)

		// Module with replacement
		require.Equal(t, "github.com/briandowns/spinner", bi.Deps[1].Path)
		require.Equal(t, "v1.11.1", bi.Deps[1].Version)
		require.Empty(t, bi.Deps[1].Sum)
		require.NotNil(t, bi.Deps[1].Replace)

		// Replacement
		require.Equal(t, "github.com/alonyb/spinner", bi.Deps[1].Replace.Path)
		require.Equal(t, "v1.12.7", bi.Deps[1].Replace.Version)
		require.Equal(t, "h1:FflTMA9I2xRd8OQ5swyZY6Q1DFeaicA/bWo6/oM82a8=", bi.Deps[1].Replace.Sum)
	})

	t.Run("Go>=1.18", func(t *testing.T) {
		out := `minikube: devel go1.18-36be0be Thu Dec 2 16:48:07 2021 +0000
path    k8s.io/minikube/cmd/minikube
mod     k8s.io/minikube (devel)
build   GOARCH=amd64
build   GOOS=linux
build   GOAMD64=v1
build   vcs=git
build   vcs.revision=febc262087e9e1f7342679bb12155072a4879316
build   vcs.time=2021-11-21T18:14:41Z
build   vcs.modified=false
`

		bi, err := parseBuildInfo("minikube", strings.NewReader(out))
		require.NoError(t, err)
		require.Equal(t, "go1.18-36be0be", bi.GoVersion)
		require.Equal(t, "k8s.io/minikube/cmd/minikube", bi.Path)
		require.NotNil(t, bi.Main)
		require.Empty(t, bi.Deps)
		require.NotNil(t, bi.Settings)

		// Main module
		require.Equal(t, "k8s.io/minikube", bi.Main.Path)
		require.Equal(t, "(devel)", bi.Main.Version)
		require.True(t, bi.Main.Main)
		require.Empty(t, bi.Main.Sum)

		// Build Settings
		require.Equal(t, "amd64", bi.Settings["GOARCH"])
		require.Equal(t, "linux", bi.Settings["GOOS"])
		require.Equal(t, "v1", bi.Settings["GOAMD64"])
		require.Equal(t, "git", bi.Settings["vcs"])
		require.Equal(t, "febc262087e9e1f7342679bb12155072a4879316", bi.Settings["vcs.revision"])
		require.Equal(t, "2021-11-21T18:14:41Z", bi.Settings["vcs.time"])
		require.Equal(t, "false", bi.Settings["vcs.modified"])
	})
}
