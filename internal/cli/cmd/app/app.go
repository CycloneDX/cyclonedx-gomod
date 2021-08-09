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

package app

import (
	"context"
	"flag"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod app", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "app",
		ShortHelp:  "Generate SBOM for an application",
		ShortUsage: "cyclonedx-gomod app [FLAGS...] PATH",
		FlagSet:    fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			cliutil.ConfigureLogger(options.LogOptions)

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	modules, err := gomod.GetModulesFromPackages(options.ModuleDir)
	if err != nil {
		return err
	}

	components, err := modconv.ToComponents(modules, modconv.WithFiles())
	if err != nil {
		return err
	}

	bom := cdx.NewBOM()
	bom.Components = &components

	return cliutil.WriteBOM(bom, options.OutputOptions)
}
