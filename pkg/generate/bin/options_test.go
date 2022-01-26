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

package bin

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestWithIncludeStdlib(t *testing.T) {
	g := &generator{includeStdlib: false}
	err := WithIncludeStdlib(true)(g)
	require.NoError(t, err)
	require.True(t, g.includeStdlib)
}

func TestWithLicenseDetection(t *testing.T) {
	g := &generator{detectLicenses: false}
	err := WithLicenseDetection(true)(g)
	require.NoError(t, err)
	require.True(t, g.detectLicenses)
}

func TestWithLogger(t *testing.T) {
	g := &generator{logger: zerolog.New(os.Stdout)}
	logger := zerolog.New(os.Stderr)
	err := WithLogger(logger)(g)
	require.NoError(t, err)
	require.Equal(t, logger, g.logger)
}

func TestWithVersionOverride(t *testing.T) {
	g := &generator{versionOverride: ""}
	err := WithVersionOverride("v1.0.0")(g)
	require.NoError(t, err)
	require.Equal(t, "v1.0.0", g.versionOverride)
}
