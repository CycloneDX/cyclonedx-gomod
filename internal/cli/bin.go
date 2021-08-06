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

package cli

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type BinOptions struct {
	OutputOptions
	SBOMOptions

	BinaryPath string
	Version    string
}

func (b *BinOptions) RegisterFlags(fs *flag.FlagSet) {
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)

	fs.StringVar(&b.Version, "version", "", "Version of the main component")
}

func (b BinOptions) Validate() error {
	if !util.FileExists(b.BinaryPath) {
		return &OptionsValidationError{Errors: []error{fmt.Errorf("binary at %s does not exist", b.BinaryPath)}}
	}

	return nil
}

func newBinCmd() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod bin", flag.ExitOnError)

	var options BinOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "bin",
		ShortHelp:  "Generate SBOM for a binary",
		ShortUsage: "cyclonedx-gomod bin [FLAGS...] PATH",
		FlagSet:    fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("no binary path provided")
			}

			options.BinaryPath = args[0]
			return execBinCmd(options)
		},
	}
}

func execBinCmd(options BinOptions) error {
	if err := options.Validate(); err != nil {
		return err
	}

	modules, hashes, err := gomod.GetModulesFromBinary(options.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to extract modules: %w", err)
	} else if len(modules) == 0 {
		return fmt.Errorf("couldn't parse any modules from %s", options.BinaryPath)
	}

	if options.Version != "" {
		modules[0].Version = options.Version
	}

	// Make all modules a direct dependency of the main module
	for i, module := range modules {
		if !module.Main {
			modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
		}
	}

	dependencies := sbom.BuildDependencyGraph(modules)

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentType(options.ComponentType)))
	if err != nil {
		return err
	}

	// Remove main module, we don't need it anymore
	modules = gomod.RemoveModule(modules, modules[0].Coordinates())

	components, err := modconv.ToComponents(modules, withModuleHashes(hashes))
	if err != nil {
		return err
	}

	compositions := make([]cdx.Composition, 0)

	// We know all components that the main component directly or indirectly depends on,
	// thus the dependencies of it are considered complete.
	compositions = append(compositions, cdx.Composition{
		Aggregate: cdx.CompositionAggregateComplete,
		Dependencies: &[]cdx.BOMReference{
			cdx.BOMReference(mainComponent.BOMRef),
		},
	})

	// The exact relationships between the dependencies are unknown
	dependencyRefs := make([]cdx.BOMReference, 0, len(components))
	for _, component := range components {
		dependencyRefs = append(dependencyRefs, cdx.BOMReference(component.BOMRef))
	}
	compositions = append(compositions, cdx.Composition{
		Aggregate:    cdx.CompositionAggregateUnknown,
		Dependencies: &dependencyRefs,
	})

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	if !options.Reproducible {
		tool, err := sbom.BuildToolMetadata()
		if err != nil {
			return fmt.Errorf("failed to build tool metadata: %w", err)
		}

		bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
		bom.Metadata.Tools = &[]cdx.Tool{*tool}
	}
	bom.Components = &components
	bom.Dependencies = &dependencies
	bom.Compositions = &compositions

	bomEncoder := cdx.NewBOMEncoder(os.Stdout, cdx.BOMFileFormatXML)
	bomEncoder.SetPretty(true)

	return bomEncoder.Encode(bom)
}

func withModuleHashes(hashes map[string]string) modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		h1, ok := hashes[m.Coordinates()]
		if !ok {
			return nil
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode h1 hash: %w", err)
		}

		c.Hashes = &[]cdx.Hash{
			{
				Algorithm: cdx.HashAlgoSHA256,
				Value:     fmt.Sprintf("%x", h1Bytes),
			},
		}

		return nil
	}
}
