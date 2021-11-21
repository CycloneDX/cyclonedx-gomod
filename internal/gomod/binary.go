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
	"io"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

func LoadModulesFromBinary(binaryPath string) (string, []Module, map[string]string, error) {
	buf := new(bytes.Buffer)
	if err := gocmd.LoadModulesFromBinary(binaryPath, buf); err != nil {
		return "", nil, nil, err
	}

	goVersion, modules, hashes := parseModulesFromBinary(binaryPath, buf)

	sortModules(modules)

	return goVersion, modules, hashes, nil
}

func parseModulesFromBinary(binaryPath string, reader io.Reader) (string, []Module, map[string]string) {
	var goVersion string
	var modules []Module
	hashes := make(map[string]string)

	moduleIndex := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		switch fields[0] {
		case binaryPath + ":":
			if len(fields) == 2 && strings.HasPrefix(fields[1], "go") {
				goVersion = fields[1]
			}
		case "mod": // Main module
			modules = append(modules, Module{
				Path:    fields[1],
				Version: fields[2],
				Main:    true,
			})
			moduleIndex += 1
		case "dep": // Dependency module
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			modules = append(modules, module)
			if len(fields) == 4 {
				// Hash won't be available when the module is replaced
				hashes[module.Coordinates()] = fields[3]
			}
			moduleIndex += 1
		case "=>": // Replacement
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			modules[moduleIndex-1].Replace = &module
			if len(fields) == 4 {
				hashes[module.Coordinates()] = fields[3]
			}
		}
	}

	return goVersion, modules, hashes
}
