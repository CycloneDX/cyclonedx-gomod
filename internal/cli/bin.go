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
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type BinOptions struct {
	OutputOptions
	SBOMOptions

	BinaryPath string
}

func (b *BinOptions) RegisterFlags(fs *flag.FlagSet) {
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)
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

	buf := new(bytes.Buffer)
	if err := gocmd.GetModulesFromBinary(options.BinaryPath, buf); err != nil {
		return err
	}

	modules := make([]gomod.Module, 0)
	hashes := make(map[string]string)

	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		switch fields[0] {
		case options.BinaryPath:
			continue
		case "path":
			continue
		case "mod":
			modules = append(modules, gomod.Module{
				Path:    fields[1],
				Version: fields[2],
				Main:    true,
			})
		case "dep":
			module := gomod.Module{
				Path:    fields[1],
				Version: fields[2],
			}
			modules = append(modules, module)
			hashes[module.Coordinates()] = fields[3]
		default:
			break
		}
	}

	if len(modules) == 0 {
		return fmt.Errorf("couldn't parse any modules from %s", options.BinaryPath)
	}

	// Make all modules a direct dependency of the main module
	for i := range modules {
		modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
	}
	dependencies := sbom.BuildDependencyGraph(modules)

	mainComponent, err := convert.ToComponent(modules[0],
		convert.WithComponentType(cdx.ComponentType(options.ComponentType)))
	if err != nil {
		return err
	}

	components, err := convert.ToComponents(modules, withModuleHashes(hashes))
	if err != nil {
		return err
	}

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	bom.Components = &components
	bom.Dependencies = &dependencies

	bomEncoder := cdx.NewBOMEncoder(os.Stdout, cdx.BOMFileFormatXML)
	bomEncoder.SetPretty(true)
	return bomEncoder.Encode(bom)
}

func withModuleHashes(hashes map[string]string) convert.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		checksum, ok := hashes[m.Coordinates()]
		if ok {
			c.Hashes = &[]cdx.Hash{
				{
					Algorithm: cdx.HashAlgoSHA256,
					Value:     checksum,
				},
			}
		}
		return nil
	}
}
