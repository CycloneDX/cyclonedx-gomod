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
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

// IsVendoring determines whether of not the module at moduleDir is vendoring its dependencies.
func IsVendoring(moduleDir string) bool {
	return util.FileExists(filepath.Join(moduleDir, "vendor", "modules.txt"))
}

var ErrNotVendoring = errors.New("the module is not vendoring its dependencies")

func GetVendoredModules(moduleDir string) ([]Module, error) {
	if !IsModule(moduleDir) {
		return nil, ErrNoModule
	}
	if !IsVendoring(moduleDir) {
		return nil, ErrNotVendoring
	}

	buf := new(bytes.Buffer)

	err := gocmd.ListVendoredModules(moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing vendored modules failed: %w", err)
	}

	modules, err := parseVendoredModules(moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("parsing vendored modules failed: %w", err)
	}

	// Main module is not included in vendored module list,
	// so we have to get it separately
	buf.Reset()
	err = gocmd.GetModule(moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing main module failed: %w", err)
	}

	var mainModule Module
	err = json.NewDecoder(buf).Decode(&mainModule)
	if err != nil {
		return nil, fmt.Errorf("parsing main module failed: %w", err)
	}

	modules = append(modules, mainModule)

	err = resolveLocalModules(moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local modules")
	}

	sortModules(modules)

	return modules, nil
}

// parseVendoredModules parses the output of `go mod vendor -v` into a Module slice.
func parseVendoredModules(mainModulePath string, reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "# ") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, "# "))

		// Replacements may be specified as
		//   Path [Version] => Path [Version]
		arrowIndex := util.StringSliceIndex(fields, "=>")

		if arrowIndex == -1 {
			if len(fields) != 2 {
				return nil, fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
			}

			modules = append(modules, Module{
				Path:     fields[0],
				Version:  fields[1],
				Dir:      filepath.Join(mainModulePath, "vendor", fields[0]),
				Vendored: true,
			})
		} else {
			pathParent := fields[0]
			versionParent := ""
			if arrowIndex == 2 {
				versionParent = fields[1]
			}

			pathReplacement := fields[arrowIndex+1]
			versionReplacement := ""
			if len(fields) == arrowIndex+3 {
				versionReplacement = fields[arrowIndex+2]
			}

			modules = append(modules, Module{
				Path:    pathParent,
				Version: versionParent,
				Replace: &Module{
					Path:     pathReplacement,
					Version:  versionReplacement,
					Dir:      filepath.Join(mainModulePath, "vendor", pathParent), // Replacements are copied to their parents dir
					Vendored: true,
				},
			})
		}
	}

	return modules, nil
}
