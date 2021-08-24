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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/rs/zerolog/log"
)

// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
type Package struct {
	Dir        string  // directory containing package sources
	ImportPath string  // import path of package in dir
	Name       string  // package name
	Standard   bool    // is this package part of the standard Go library?
	Module     *Module // info about package's containing module, if any (can be nil)

	GoFiles        []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles       []string // .go source files that import "C"
	CFiles         []string // .c source files
	CXXFiles       []string // .cc, .cxx and .cpp source files
	MFiles         []string // .m source files
	HFiles         []string // .h, .hh, .hpp and .hxx source files
	FFiles         []string // .f, .F, .for and .f90 Fortran source files
	SFiles         []string // .s source files
	SwigFiles      []string // .swig files
	SwigCXXFiles   []string // .swigcxx files
	SysoFiles      []string // .syso object files to add to archive
	TestGoFiles    []string // _test.go files in package
	EmbedFiles     []string // files matched by EmbedPatterns
	TestEmbedFiles []string // files matched by TestEmbedPatterns
}

func GetModulesFromPackages(moduleDir, packagePattern string) ([]Module, error) {
	buf := new(bytes.Buffer)

	err := gocmd.ListPackages(moduleDir, packagePattern, buf)
	if err != nil {
		return nil, err
	}

	pkgMap, err := parsePackages(buf)
	if err != nil {
		return nil, err
	}

	modules, err := convertPackages(pkgMap)
	if err != nil {
		return nil, err
	}

	sortModules(modules)

	return modules, nil
}

// parsePackageInfo parses the output of `go list -json`.
// The keys of the returned map are module coordinates (path@version).
func parsePackages(reader io.Reader) (map[string][]Package, error) {
	pkgsMap := make(map[string][]Package)
	jsonDecoder := json.NewDecoder(reader)

	for {
		var pkg Package
		if err := jsonDecoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		if pkg.Standard {
			log.Debug().
				Str("package", pkg.ImportPath).
				Msg("skipping standard library package")
			continue
		}
		if pkg.Module == nil {
			log.Debug().
				Str("package", pkg.ImportPath).
				Msg("skipping package without module")
			continue
		}

		pkgs, ok := pkgsMap[pkg.Module.Coordinates()]
		if !ok {
			pkgsMap[pkg.Module.Coordinates()] = []Package{pkg}
		} else {
			pkgsMap[pkg.Module.Coordinates()] = append(pkgs, pkg)
		}
	}

	return pkgsMap, nil
}

func convertPackages(pkgsMap map[string][]Package) ([]Module, error) {
	modules := make([]Module, 0, len(pkgsMap))

	for _, pkgs := range pkgsMap {
		var module *Module
		if len(pkgs) > 0 {
			module = pkgs[0].Module
			if module == nil {
				// Shouldn't ever happen, because packages without module are not collected to pkgsMap.
				// We do the nil check anyway to make linters happy. :)
				return nil, fmt.Errorf("no module is associated with package %s", pkgs[0].ImportPath)
			}
		} else {
			continue
		}

		for i := range pkgs {
			var pkgFiles []string
			pkgFiles = append(pkgFiles, pkgs[i].GoFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].CgoFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].CFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].CXXFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].MFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].HFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].FFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].SFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].SwigFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].SwigCXXFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].SysoFiles...)
			pkgFiles = append(pkgFiles, pkgs[i].EmbedFiles...)

			for _, file := range pkgFiles {
				if filePath, err := makePackageFileRelativeToModule(module.Dir, pkgs[i].Dir, file); err == nil {
					module.Files = append(module.Files, filePath)
				} else {
					return nil, err
				}
			}

			var pkgTestFiles []string
			pkgTestFiles = append(pkgTestFiles, pkgs[i].TestGoFiles...)
			pkgTestFiles = append(pkgTestFiles, pkgs[i].TestEmbedFiles...)

			for _, testFile := range pkgTestFiles {
				if testFilePath, err := makePackageFileRelativeToModule(module.Dir, pkgs[i].Dir, testFile); err == nil {
					module.TestFiles = append(module.TestFiles, testFilePath)
				} else {
					return nil, err
				}
			}

			sort.Slice(module.Files, func(i, j int) bool {
				return module.Files[i] < module.Files[j]
			})
			sort.Slice(module.TestFiles, func(i, j int) bool {
				return module.TestFiles[i] < module.TestFiles[j]
			})
		}

		modules = append(modules, *module)
	}

	return modules, nil
}

func makePackageFileRelativeToModule(moduleDir, pkgDir, file string) (string, error) {
	fp, err := filepath.Rel(moduleDir, filepath.Join(pkgDir, file))
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(fp, "\\", "/"), nil
}
