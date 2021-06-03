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

package gomod

import (
	"encoding/json"
	"errors"
	"io"
	"log"
)

// PackageInfo represents parts of the struct that `go list` is working with.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
type PackageInfo struct {
	Dir         string   // directory containing package sources
	ImportPath  string   // import path of package in dir
	Name        string   // package name
	Standard    bool     // is this package part of the standard Go library?
	Module      *Module  // info about package's containing module, if any (can be nil)
	GoFiles     []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	TestGoFiles []string // _test.go files in package
}

// parsePackageInfo parses the output of `go list -json`.
//
// The keys of the returned map are module coordinates (path@version).
func parsePackageInfo(reader io.Reader) (map[string][]PackageInfo, error) {
	pkgsMap := make(map[string][]PackageInfo, 0)
	jsonDecoder := json.NewDecoder(reader)

	for {
		var pkg PackageInfo
		if err := jsonDecoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if pkg.Standard {
			log.Printf("skipping stdlib package %s\n", pkg.ImportPath)
			continue
		}
		if pkg.Module == nil {
			log.Printf("skipping package without module %s\n", pkg.ImportPath)
			continue
		}

		pkgs, ok := pkgsMap[pkg.Module.Coordinates()]
		if !ok {
			pkgsMap[pkg.Module.Coordinates()] = []PackageInfo{pkg}
		} else {
			pkgsMap[pkg.Module.Coordinates()] = append(pkgs, pkg)
		}
	}

	return pkgsMap, nil
}
