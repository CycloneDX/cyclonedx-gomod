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

package bin

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type BinOptions struct {
	options.OutputOptions
	options.SBOMOptions

	BinaryPath string
	Version    string
}

func (b *BinOptions) RegisterFlags(fs *flag.FlagSet) {
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)

	fs.StringVar(&b.Version, "version", "", "Version of the main component")
}

func (b BinOptions) Validate() error {
	errs := make([]error, 0)

	if err := b.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := b.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	if !util.FileExists(b.BinaryPath) {
		errs = append(errs, fmt.Errorf("binary at %s does not exist", b.BinaryPath))
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}

func New() *ffcli.Command {
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

func execBinCmd(binOptions BinOptions) error {
	if err := binOptions.Validate(); err != nil {
		return err
	}

	modules, hashes, err := gomod.GetModulesFromBinary(binOptions.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to extract modules: %w", err)
	} else if len(modules) == 0 {
		return fmt.Errorf("couldn't parse any modules from %s", binOptions.BinaryPath)
	}

	if binOptions.Version != "" {
		modules[0].Version = binOptions.Version
	}

	// Make all modules a direct dependency of the main module
	for i, module := range modules {
		if !module.Main {
			modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
		}
	}

	dependencies := sbom.BuildDependencyGraph(modules)

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentType(binOptions.ComponentType)),
		modconv.WithScope(""), // Main component can't have a scope
	)
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

	binaryHashes, err := sbom.CalculateFileHashes(binOptions.BinaryPath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return fmt.Errorf("failed to calculate binary hashes: %w", err)
	}

	properties := []cdx.Property{
		sbom.NewProperty("binary:name", filepath.Base(binOptions.BinaryPath)),
	}
	for _, hash := range binaryHashes {
		properties = append(properties, sbom.NewProperty(fmt.Sprintf("binary:hash:%s", hash.Algorithm), hash.Value))
	}

	bom := cdx.NewBOM()

	if err = cliutil.SetSerialNumber(bom, binOptions.SBOMOptions); err != nil {
		return err
	}

	bom.Metadata = &cdx.Metadata{
		Component:  mainComponent,
		Properties: &properties,
	}
	if !binOptions.Reproducible {
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

	return cliutil.WriteBOM(bom, binOptions.OutputOptions)
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
