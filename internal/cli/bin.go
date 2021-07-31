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
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/model"
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

	modules := make([]model.Module, 0)

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
			modules = append(modules, model.Module{
				Path:    fields[1],
				Version: fields[2],
			})
		case "dep":
			modules = append(modules, model.Module{
				Path:     fields[1],
				Version:  fields[2],
				Checksum: fields[3],
			})
		default:
			break
		}
	}

	if len(modules) == 0 {
		return fmt.Errorf("couldn't parse any modules from %s", options.BinaryPath)
	}

	return nil
}
