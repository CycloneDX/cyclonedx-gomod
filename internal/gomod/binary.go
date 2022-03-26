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
	"debug/buildinfo"
	"fmt"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// BuildInfo represents the build information read from a Go binary.
// Adapted from https://github.com/golang/go/blob/931d80ec17374e52dbc5f9f63120f8deb80b355d/src/runtime/debug/mod.go#L41
type BuildInfo struct {
	GoVersion string            // Version of Go that produced this binary.
	Path      string            // The main package path
	Main      *Module           // The module containing the main package
	Deps      []Module          // Module dependencies
	Settings  map[string]string // Other information about the build.
}

func LoadBuildInfo(binaryPath string) (*BuildInfo, error) {
	stdBuildInfo, err := buildinfo.ReadFile(binaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read build info: %w", err)
	}

	buildInfo := BuildInfo{
		Path: stdBuildInfo.Path,
		Main: &Module{
			Path:    stdBuildInfo.Main.Path,
			Version: stdBuildInfo.Main.Version,
			Main:    true,
			Sum:     stdBuildInfo.Main.Sum,
		},
	}

	buildInfo.GoVersion, err = gocmd.ParseVersion(stdBuildInfo.GoVersion)
	if err != nil {
		return nil, err
	}

	var deps []Module
	for i := range stdBuildInfo.Deps {
		dep := Module{
			Path:    stdBuildInfo.Deps[i].Path,
			Version: stdBuildInfo.Deps[i].Version,
			Sum:     stdBuildInfo.Deps[i].Sum,
		}
		if stdBuildInfo.Deps[i].Replace != nil {
			dep.Replace = &Module{
				Path:    stdBuildInfo.Deps[i].Replace.Path,
				Version: stdBuildInfo.Deps[i].Replace.Version,
				Sum:     stdBuildInfo.Deps[i].Replace.Sum,
			}
		}
		deps = append(deps, dep)
	}
	if len(deps) > 0 {
		// Make all deps a direct dependency of main
		buildInfo.Main.Dependencies = make([]*Module, len(deps))
		for i := range deps {
			buildInfo.Main.Dependencies[i] = &deps[i]
		}
		sortDependencies(buildInfo.Main.Dependencies)

		buildInfo.Deps = deps
	}

	settings := make(map[string]string)
	for _, setting := range stdBuildInfo.Settings {
		settings[setting.Key] = setting.Value
	}
	if len(settings) > 0 {
		buildInfo.Settings = settings
	}

	return &buildInfo, nil
}
