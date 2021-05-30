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
// Copyright (c) Niklas Düster. All Rights Reserved.

package util

import (
	"go/build"
	"os"
	"path/filepath"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsGoModule determines whether the directory at the given path is a Go module.
func IsGoModule(path string) bool {
	return FileExists(filepath.Join(path, "go.mod"))
}

// IsVendoring determines whether of not the module at the given path is vendoring its dependencies.
// Should be used in conjunction with IsGoModule.
func IsVendoring(path string) bool {
	return FileExists(filepath.Join(path, "vendor", "modules.txt"))
}

// GetGoPath determines the GOPATH location.
func GetGoPath() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	return gopath
}

// GetModuleCacheDir determines the location of Go's module cache.
func GetModuleCacheDir() string {
	modCacheDir := os.Getenv("GOMODCACHE")
	if modCacheDir == "" {
		modCacheDir = filepath.Join(GetGoPath(), "pkg", "mod")
	}
	return modCacheDir
}

// StringSliceIndex determines the index of a given string in a given string slice.
func StringSliceIndex(haystack []string, needle string) int {
	for i := range haystack {
		if haystack[i] == needle {
			return i
		}
	}
	return -1
}
