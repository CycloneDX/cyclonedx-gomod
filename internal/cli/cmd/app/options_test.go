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
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/stretchr/testify/require"
)

func TestOptions_Validate(t *testing.T) {
	t.Run("Files without Packages", func(t *testing.T) {
		var options Options
		options.IncludeFiles = true
		options.IncludePackages = false

		err := options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not supported")
	})

	t.Run("Main Isnt Subpath Of MODULE_PATH", func(t *testing.T) {
		var options Options
		options.ModuleDir = "/path/to/module"
		options.Main = "../main.go"

		err := options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be a subpath")
	})

	t.Run("Main Doesnt Exist", func(t *testing.T) {
		var options Options
		options.ModuleDir = "/path/to/module"
		options.Main = "cmd/app"

		err := options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not exist")
	})

	t.Run("Main Is File", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0600)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "main.go"

		err = options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be a directory")
	})

	t.Run("Main Isnt A Main Package", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.MkdirAll(filepath.Join(tmpDir, "cmd/app"), 0700)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module foobar"), 0600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "cmd/app/main.go"), []byte("package baz"), 0600)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "cmd/app"

		err = options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be main package")
	})

	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.MkdirAll(filepath.Join(tmpDir, "cmd/app"), 0700)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module foobar"), 0600)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(tmpDir, "cmd/app/main.go"), []byte("package main"), 0600)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "cmd/app"
		options.OutputVersion = cdx.SpecVersion1_4.String()

		err = options.Validate()
		require.NoError(t, err)
	})
}
