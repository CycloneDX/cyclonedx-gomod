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

package app

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect/standard"
)

func TestWithIncludeFiles(t *testing.T) {
	g := &generator{includeFiles: false}
	err := WithIncludeFiles(true)(g)
	require.NoError(t, err)
	require.True(t, g.includeFiles)
}

func TestWithIncludePackages(t *testing.T) {
	g := &generator{includePackages: false}
	err := WithIncludePackages(true)(g)
	require.NoError(t, err)
	require.True(t, g.includePackages)
}

func TestWithIncludeStdlib(t *testing.T) {
	g := &generator{includeStdlib: false}
	err := WithIncludeStdlib(true)(g)
	require.NoError(t, err)
	require.True(t, g.includeStdlib)
}

func TestWithLicenseDetector(t *testing.T) {
	detector := standard.NewDetector(zerolog.Nop())

	g := &generator{licenseDetector: detector}
	err := WithLicenseDetector(detector)(g)
	require.NoError(t, err)
	require.Equal(t, detector, g.licenseDetector)
}

func TestWithLogger(t *testing.T) {
	g := &generator{logger: zerolog.New(os.Stdout)}
	logger := zerolog.New(os.Stderr)
	err := WithLogger(logger)(g)
	require.NoError(t, err)
	require.Equal(t, logger, g.logger)
}
