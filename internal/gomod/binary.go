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

func GetModulesFromBinary(binaryPath string) ([]Module, map[string]string, error) {
	buf := new(bytes.Buffer)
	if err := gocmd.GetModulesFromBinary(binaryPath, buf); err != nil {
		return nil, nil, err
	}

	modules, hashes := parseModulesFromBinary(buf)

	sortModules(modules)

	return modules, hashes, nil
}

func parseModulesFromBinary(reader io.Reader) ([]Module, map[string]string) {
	modules := make([]Module, 0)
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
			hashes[module.Coordinates()] = fields[3]
		}
	}

	return modules, hashes
}
