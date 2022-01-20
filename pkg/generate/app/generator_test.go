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
	"errors"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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

func TestGenerator_CreateBuildProperties(t *testing.T) {
	g := generator{
		logger: zerolog.New(io.Discard),
	}

	origGoflags := os.Getenv("GOFLAGS")
	os.Setenv("GOFLAGS", "-tags=foo,bar")

	if origGoflags != "" {
		defer func() {
			os.Setenv("GOFLAGS", origGoflags)
		}()
	}

	properties, err := g.createBuildProperties()
	require.NoError(t, err)
	require.Len(t, properties, 6)

	expectedCgoEnabled := "1" // Cgo is enabled per default
	if cgo := os.Getenv("CGO_ENABLED"); cgo != "" {
		expectedCgoEnabled = cgo
	}

	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:CGO_ENABLED", Value: expectedCgoEnabled})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOARCH", Value: runtime.GOARCH})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOOS", Value: runtime.GOOS})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:env:GOVERSION", Value: runtime.Version()})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "foo"})
	assert.Contains(t, properties, cyclonedx.Property{Name: "cdx:gomod:build:tag", Value: "bar"})
}
