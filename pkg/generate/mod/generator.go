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

package mod

import (
	"errors"
	"fmt"
	"io"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
)

type generator struct {
	logger zerolog.Logger

	moduleDir       string
	componentType   cdx.ComponentType
	includeStdlib   bool
	includeTest     bool
	licenseDetector licensedetect.Detector
}

// NewGenerator returns a generator that is capable of generating BOMs for Go modules.
func NewGenerator(moduleDir string, opts ...Option) (generate.Generator, error) {
	g := generator{
		logger:        log.Logger,
		moduleDir:     moduleDir,
		componentType: cdx.ComponentTypeApplication,
	}

	var err error
	for _, opt := range opts {
		if err = opt(&g); err != nil {
			return nil, err
		}
	}

	return &g, nil
}

// Generate implements the generate.Generator interface.
func (g generator) Generate() (*cdx.BOM, error) {
	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	err := gocmd.ModWhy(g.logger, g.moduleDir, []string{"github.com/CycloneDX/cyclonedx-go"}, io.Discard)
	if err != nil {
		return nil, fmt.Errorf("failed to download modules: %w", err)
	}

	modules, err := gomod.GetVendoredModules(g.logger, g.moduleDir, g.includeTest)
	if err != nil {
		if errors.Is(err, gomod.ErrNotVendoring) {
			modules, err = gomod.LoadModules(g.logger, g.moduleDir, g.includeTest)
			if err != nil {
				return nil, fmt.Errorf("failed to collect modules: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to collect vendored modules: %w", err)
		}
	}

	if g.includeStdlib {
		stdlibModule, err := gomod.LoadStdlibModule(g.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to load stdlib module: %w", err)
		}

		modules[0].Dependencies = append(modules[0].Dependencies, stdlibModule)
		modules = append(modules, *stdlibModule)
	}

	err = gomod.ApplyModuleGraph(g.logger, g.moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to apply module graph: %w", err)
	}

	modules[0].Version, err = gomod.GetModuleVersion(g.logger, modules[0].Dir)
	if err != nil {
		g.logger.Warn().Err(err).Msg("failed to determine version of main module")
	}

	main, err := modConv.ToComponent(g.logger, modules[0],
		modConv.WithComponentType(g.componentType),
		modConv.WithLicenses(g.licenseDetector),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}
	components, err := modConv.ToComponents(g.logger, modules[1:],
		modConv.WithLicenses(g.licenseDetector),
		modConv.WithModuleHashes(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert modules: %w", err)
	}
	dependencies := sbom.BuildDependencyGraph(modules)

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: main,
	}
	bom.Components = &components
	bom.Dependencies = &dependencies

	return bom, nil
}
