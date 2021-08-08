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
	"flag"
	"fmt"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod bin", flag.ExitOnError)

	var options BinOptions
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "bin",
		ShortHelp:  "Generate SBOM for a binary",
		ShortUsage: "cyclonedx-gomod bin [FLAGS...] PATH",
		LongHelp: `Generate SBOM for a binary.

Please note that data embedded in binaries shouldn't be trusted,
unless there's solid evidence that the binaries haven't been modified
since they've been built.

Example:
  $ cyclonedx-gomod bin -json -output minikube-v1.22.0.bom.json -version v1.22.0 ./minikube`,
		FlagSet: fs,
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
	err := binOptions.Validate()
	if err != nil {
		return err
	}

	cliutil.ConfigureLogger(binOptions.LogOptions)

	modules, hashes, err := gomod.GetModulesFromBinary(binOptions.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to extract modules: %w", err)
	} else if len(modules) == 0 {
		return fmt.Errorf("couldn't parse any modules from %s", binOptions.BinaryPath)
	}

	if binOptions.Version != "" {
		modules[0].Version = binOptions.Version
	}

	sbom.NormalizeVersions(modules, binOptions.NoVersionPrefix)

	// Make all modules a direct dependency of the main module
	for i := range modules[1:] {
		modules[0].Dependencies = append(modules[0].Dependencies, &modules[i+1])
	}

	dependencies := sbom.BuildDependencyGraph(modules)

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentTypeApplication),
		modconv.WithScope(""), // Main component can't have a scope
	)
	if err != nil {
		return err
	}

	components, err := modconv.ToComponents(modules[1:], withModuleHashes(hashes))
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

	binaryProperties, err := createBinaryProperties(binOptions.BinaryPath)
	if err != nil {
		return err
	}

	bom := cdx.NewBOM()

	if err = cliutil.SetSerialNumber(bom, binOptions.SBOMOptions); err != nil {
		return err
	}

	bom.Metadata = &cdx.Metadata{
		Component:  mainComponent,
		Properties: &binaryProperties,
	}
	if err = cliutil.AddCommonMetadata(bom, binOptions.SBOMOptions); err != nil {
		return err
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

func createBinaryProperties(binaryPath string) ([]cdx.Property, error) {
	binaryHashes, err := sbom.CalculateFileHashes(binaryPath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate binary hashes: %w", err)
	}

	properties := []cdx.Property{
		sbom.NewProperty("binary:name", filepath.Base(binaryPath)),
	}
	for _, hash := range binaryHashes {
		properties = append(properties, sbom.NewProperty(fmt.Sprintf("binary:hash:%s", hash.Algorithm), hash.Value))
	}

	return properties, nil
}
