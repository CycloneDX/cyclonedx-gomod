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
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// ModOptions provides options for the `mod` command.
type ModOptions struct {
	options.OutputOptions
	options.SBOMOptions

	ModuleDir       string
	ResolveLicenses bool
}

func (m *ModOptions) RegisterFlags(fs *flag.FlagSet) {
	m.OutputOptions.RegisterFlags(fs)
	m.SBOMOptions.RegisterFlags(fs)

	fs.BoolVar(&m.ResolveLicenses, "licenses", false, "Resolve module licenses")
}

func (m ModOptions) Validate() error {
	errs := make([]error, 0)

	if err := m.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := m.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)

	var options ModOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOM for a module",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] [PATH]",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			return execModCmd(options)
		},
	}
}

func execModCmd(modOptions ModOptions) error {
	if err := modOptions.Validate(); err != nil {
		return err
	}

	var serial *uuid.UUID
	if !modOptions.NoSerialNumber && modOptions.SerialNumber != "" {
		serialUUID := uuid.MustParse(modOptions.SerialNumber)
		serial = &serialUUID
	}

	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	log.Println("downloading modules")
	if err := gocmd.ModWhy(modOptions.ModuleDir, []string{"github.com/CycloneDX/cyclonedx-go"}, io.Discard); err != nil {
		return fmt.Errorf("downloading modules failed: %w", err)
	}

	log.Println("enumerating modules")
	modules, err := gomod.GetModules(modOptions.ModuleDir, modOptions.IncludeTest)
	if err != nil {
		return fmt.Errorf("failed to enumerate modules: %w", err)
	}

	log.Println("normalizing module versions")
	for i := range modules {
		modules[i].Version = strings.TrimSuffix(modules[i].Version, "+incompatible")

		if modOptions.NoVersionPrefix {
			modules[i].Version = strings.TrimPrefix(modules[i].Version, "v")
		}
	}

	mainModule := modules[0]
	modules = modules[1:]

	log.Println("determining version of main module")
	if mainModule.Version, err = gomod.GetModuleVersion(mainModule.Dir); err != nil {
		log.Printf("failed to get version of main module: %v\n", err)
	}
	if mainModule.Version != "" && modOptions.NoVersionPrefix {
		mainModule.Version = strings.TrimPrefix(mainModule.Version, "v")
	}

	log.Printf("converting main module %s\n", mainModule.Coordinates())
	mainComponent, err := modconv.ToComponent(mainModule,
		modconv.WithComponentType(cdx.ComponentType(modOptions.ComponentType)),
		withLicenses(modOptions.ResolveLicenses),
		modconv.WithScope(""), // Main component can't have a scope
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	components, err := modconv.ToComponents(modules,
		withModuleHashes(),
		withLicenses(modOptions.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	log.Println("building dependency graph")
	dependencyGraph := sbom.BuildDependencyGraph(append(modules, mainModule))

	log.Println("assembling sbom")
	bom := cdx.NewBOM()
	if !modOptions.NoSerialNumber {
		if serial == nil {
			bom.SerialNumber = uuid.New().URN()
		} else {
			bom.SerialNumber = serial.URN()
		}
	}

	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	if !modOptions.Reproducible {
		tool, err := sbom.BuildToolMetadata()
		if err != nil {
			return fmt.Errorf("failed to build tool metadata: %w", err)
		}

		bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
		bom.Metadata.Tools = &[]cdx.Tool{*tool}
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if modOptions.IncludeStd {
		log.Println("gathering info about standard library")
		stdComponent, err := sbom.BuildStdComponent()
		if err != nil {
			return fmt.Errorf("failed to build std component: %w", err)
		}

		log.Println("adding standard library to sbom")
		*bom.Components = append(*bom.Components, *stdComponent)

		// Add std to dependency graph
		stdDependency := cdx.Dependency{Ref: stdComponent.BOMRef}
		*bom.Dependencies = append(*bom.Dependencies, stdDependency)

		// Add std as dependency of main module
		for i, dependency := range *bom.Dependencies {
			if dependency.Ref == mainComponent.BOMRef {
				if dependency.Dependencies == nil {
					(*bom.Dependencies)[i].Dependencies = &[]cdx.Dependency{stdDependency}
				} else {
					*dependency.Dependencies = append(*dependency.Dependencies, stdDependency)
				}
				break
			}
		}
	}

	return cliutil.WriteBOM(bom, modOptions.OutputOptions)
}

func withLicenses(enabled bool) modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if enabled {
			return modconv.WithLicenses()(m, c)
		}
		return nil
	}
}

func withModuleHashes() modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if m.Main || m.Vendored {
			return nil
		}

		h1, err := m.Hash()
		if err != nil {
			return fmt.Errorf("failed to calculate h1 hash: %w", err)
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode h1 hash: %w", err)
		}

		c.Hashes = &[]cdx.Hash{
			{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
		}

		return nil
	}
}
