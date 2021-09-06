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

	"github.com/stretchr/testify/require"
)

func TestOptions_Validate(t *testing.T) {
	t.Run("Main Isnt A Go File", func(t *testing.T) {
		var options Options
		options.Main = "./notGo.txt"

		err := options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "must be a go source file")
	})

	t.Run("Main Isnt Subpath Of MODDIR", func(t *testing.T) {
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
		options.Main = "main.go"

		err := options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not exist")
	})

	t.Run("Main Is Directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.Mkdir(filepath.Join(tmpDir, "main.go"), os.ModePerm)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "main.go"

		err = options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "is a directory")
	})

	t.Run("Main Isnt A Main File", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package notmain`), os.ModePerm)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "main.go"

		err = options.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not a main file")
	})

	t.Run("Success", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package  main // somecomment`), os.ModePerm)
		require.NoError(t, err)

		var options Options
		options.ModuleDir = tmpDir
		options.Main = "main.go"

		err = options.Validate()
		require.NoError(t, err)
	})
}
