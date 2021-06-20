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
	"path/filepath"
	"strings"
)

// Package represents parts of the struct that `go list` is working with.
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

// parsePackageInfo parses the output of `go list -json`.
//
// The keys of the returned map are module coordinates (path@version).
func parsePackages(reader io.Reader) (map[string][]Package, error) {
	pkgsMap := make(map[string][]Package, 0)
	jsonDecoder := json.NewDecoder(reader)

	for {
		var pkg Package
		if err := jsonDecoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		if pkg.Standard || pkg.Module == nil {
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

		for i := range pkgs {
			if module == nil {
				module = pkgs[i].Module
			}

			pkgFiles := make([]string, 0)
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

			pkgTestFiles := make([]string, 0)
			pkgTestFiles = append(pkgTestFiles, pkgs[i].TestGoFiles...)
			pkgTestFiles = append(pkgTestFiles, pkgs[i].TestEmbedFiles...)

			for _, testFile := range pkgTestFiles {
				if testFilePath, err := makePackageFileRelativeToModule(module.Dir, pkgs[i].Dir, testFile); err == nil {
					module.TestFiles = append(module.TestFiles, testFilePath)
				} else {
					return nil, err
				}
			}
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
