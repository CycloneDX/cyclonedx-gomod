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

	GoFiles      []string // .go source files (excluding CgoFiles, TestGoFiles, XTestGoFiles)
	CgoFiles     []string // .go source files that import "C"
	CFiles       []string // .c source files
	CXXFiles     []string // .cc, .cxx and .cpp source files
	MFiles       []string // .m source files
	HFiles       []string // .h, .hh, .hpp and .hxx source files
	FFiles       []string // .f, .F, .for and .f90 Fortran source files
	SFiles       []string // .s source files
	SwigFiles    []string // .swig files
	SwigCXXFiles []string // .swigcxx files
	SysoFiles    []string // .syso object files to add to archive
	EmbedFiles   []string // files matched by EmbedPatterns

	Error *PackageError // error loading package
}

type PackageError struct {
	Err string
}

func (pe PackageError) Error() string {
	return pe.Err
}

func LoadPackage(moduleDir, packagePattern string) (*Package, error) {
	log.Debug().
		Str("moduleDir", moduleDir).
		Str("packagePattern", packagePattern).
		Msg("loading package")

	buf := new(bytes.Buffer)
	err := gocmd.ListPackage(moduleDir, toRelativePackagePath(packagePattern), buf)
	if err != nil {
		return nil, err
	}

	var pkg Package
	err = json.NewDecoder(buf).Decode(&pkg)
	if err != nil {
		return nil, err
	}

	if pkg.Error != nil {
		return nil, pkg.Error
	}

	return &pkg, nil
}

func LoadModulesFromPackages(moduleDir, packagePattern string) ([]Module, error) {
	log.Debug().
		Str("moduleDir", moduleDir).
		Msg("loading modules")

	if !IsModule(moduleDir) {
		return nil, ErrNoModule
	}

	buf := new(bytes.Buffer)
	err := gocmd.ListPackages(moduleDir, toRelativePackagePath(packagePattern), buf)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages for pattern \"%s\": %w", packagePattern, err)
	}

	pkgMap, err := parsePackages(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to parse `go list` output: %w", err)
	}

	modules, err := convertPackages(moduleDir, pkgMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert packages to modules: %w", err)
	}

	err = ResolveLocalReplacements(moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local replacements: %w", err)
	}

	sortModules(modules)

	return modules, nil
}

// parsePackages parses the output of `go list -json`.
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
				Str("reason", "part of standard library").
				Msg("skipping package")
			continue
		}
		if pkg.Module == nil {
			log.Debug().
				Str("package", pkg.ImportPath).
				Str("reason", "no associated module").
				Msg("skipping package")
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

func convertPackages(mainModuleDir string, pkgsMap map[string][]Package) ([]Module, error) {
	modules := make([]Module, 0, len(pkgsMap))
	vendoring := IsVendoring(mainModuleDir)

	// Collect modules
	for _, pkgs := range pkgsMap {
		if len(pkgs) == 0 {
			continue
		}

		module := pkgs[0].Module
		if module == nil {
			// Shouldn't ever happen, because packages without module are not collected to pkgsMap.
			// We do the nil check anyway to make linters happy. :)
			return nil, fmt.Errorf("no module is associated with package %s", pkgs[0].ImportPath)
		}

		if !module.Main && vendoring {
			module.Vendored = true
			vendorPath := filepath.Join(mainModuleDir, "vendor", module.Path)

			if module.Replace != nil {
				module.Replace.Vendored = true
				module.Replace.Dir = vendorPath
			} else {
				module.Dir = vendorPath
			}
		}

		modules = append(modules, *module)
	}

	// Assign packages to modules
	for i := range modules {
		pkgs := pkgsMap[modules[i].Coordinates()]
		for j := range pkgs {
			pkgs[j].Module = nil // prevent cyclic references
			modules[i].Packages = append(modules[i].Packages, pkgs[j])
		}
	}

	return modules, nil
}

// toRelativePackagePath ensures that Go will interpret the given packagePattern
// as relative package path. This is done by prefixing it with "./" if it isn't already.
// See also: `go help packages`
func toRelativePackagePath(packagePattern string) string {
	packagePattern = filepath.ToSlash(packagePattern)
	if !strings.HasPrefix(packagePattern, "./") {
		packagePattern = "./" + packagePattern
	}
	return packagePattern
}
