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
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliutil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod app", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "app",
		ShortHelp:  "Generate SBOM for an application",
		ShortUsage: "cyclonedx-gomod app [FLAGS...] MODPATH",
		LongHelp: `Generate SBOM for an application.

In order to produce accurate results, build constraints must be configured
via environment variables. These build constraints should mimic the ones passed
to the "go build" command for the application.

Noteworthy environment variables that act as build constraints are:
  - GOARCH       The target architecture (386, amd64, etc.)
  - GOOS         The target operating system (linux, windows, etc.)
  - CGO_ENABLED  Whether or not CGO is enabled
  - GOFLAGS      Pass build tags

A complete overview of all environment variables can be found here:
  https://pkg.go.dev/cmd/go#hdr-Environment_variables

Unless the -reproducible flag is provided, build constraints will be 
included as properties of the main component.

The -main flag should be used to specify the path to the application's main file.
-main must point to a go file within MODPATH. If -main is not specified, "main.go" is assumed.

By passing -files, all files that would be compiled into the binary will be included
as subcomponents of their respective module. Files versions follow the v0.0.0-SHORTHASH pattern, 
where SHORTHASH is the first 12 characters of the file's SHA1 hash.

Examples:
  $ GOARCH=arm64 GOOS=linux GOFLAGS="-tags=foo,bar" cyclonedx-gomod app -output linux-arm64.bom.xml
  $ cyclonedx-gomod app -json -output acme-app.bom.json -files -licenses -main cmd/acme-app/main.go /usr/src/acme-module`,
		FlagSet: fs,
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

	modules, err := gomod.GetModulesFromPackages(options.ModuleDir, options.Main)
	if err != nil {
		return err
	}

	// Dependencies need to be applied prior to determining the main
	// module's version, because `go mod graph` omits that version.
	err = gomod.ApplyModuleGraph(options.ModuleDir, modules)
	if err != nil {
		return err
	}

	modules[0].Version, err = gomod.GetModuleVersion(modules[0].Dir)
	if err != nil {
		return err
	}

	mainComponent, err := modconv.ToComponent(modules[0],
		modconv.WithComponentType(cdx.ComponentTypeApplication),
		modconv.WithFiles(options.IncludeFiles),
		modconv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return err
	}

	if !options.SBOMOptions.Reproducible {
		buildProperties, err := createBuildProperties()
		if err != nil {
			return err
		}
		if mainComponent.Properties == nil {
			mainComponent.Properties = &buildProperties
		} else {
			*mainComponent.Properties = append(*mainComponent.Properties, buildProperties...)
		}
	}

	components, err := modconv.ToComponents(modules[1:],
		modconv.WithFiles(options.IncludeFiles),
		modconv.WithLicenses(options.ResolveLicenses),
		modconv.WithModuleHashes(),
	)
	if err != nil {
		return err
	}

	dependencies := sbom.BuildDependencyGraph(modules)

	bom := cdx.NewBOM()

	err = cliutil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	err = cliutil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return err
	}

	bom.Components = &components
	bom.Dependencies = &dependencies

	enrichWithApplicationDetails(bom, options.ModuleDir, options.Main)

	if options.IncludeStd {
		err = cliutil.AddStdComponent(bom)
		if err != nil {
			return err
		}
	}

	return cliutil.WriteBOM(bom, options.OutputOptions)
}

var buildEnv = []string{
	"CGO_ENABLED",
	"GOARCH",
	"GOOS",
	"GOVERSION",
}

func createBuildProperties() (properties []cdx.Property, err error) {
	env, err := gocmd.GetEnv()
	if err != nil {
		return nil, err
	}

	for _, buildEnvKey := range buildEnv {
		buildEnvVal, ok := env[buildEnvKey]
		if !ok {
			log.Warn().
				Str("env", buildEnvKey).
				Msg("environment variable not found")
			continue
		}

		if buildEnvVal != "" {
			properties = append(properties, sbom.NewProperty("build:env:"+buildEnvKey, buildEnvVal))
		}
	}

	goflags, ok := env["GOFLAGS"]
	if ok {
		tags := parseTagsFromGoFlags(goflags)
		for _, tag := range tags {
			properties = append(properties, sbom.NewProperty("build:tag", tag))
		}
	}

	return
}

func parseTagsFromGoFlags(goflags string) (tags []string) {
	fields := strings.Fields(goflags)

	for _, field := range fields {
		if !strings.HasPrefix(field, "-tags=") {
			continue
		}

		tagList := strings.Split(field, "=")[1]
		tags = append(tags, strings.Split(tagList, ",")...)
	}

	return
}

// enrichWithApplicationDetails determines the application name as well as
// the path to the application (path to mainFile's parent dir) relative to moduleDir.
// If the application path is not equal to moduleDir, it is added to the main component's
// package URL as sub path. For example:
//
// + moduleDir <- application name
// |-+ main.go
//
// + moduleDir
// |-+ cmd
//   |-+ app   <- application name
//     |-+ main.go
//
// The package URLs for the above examples would look like this:
//   1. pkg:golang/../module@version         (untouched)
//   2. pkg:golang/../module@version#cmd/app (with sub path)
//
// If the package URL is updated, the BOM reference is as well.
// All places within the BOM that reference the main component will be updated accordingly.
func enrichWithApplicationDetails(bom *cdx.BOM, moduleDir, mainFile string) {
	// Resolve absolute paths to moduleDir and mainFile.
	// Both may contain traversals or similar elements we don't care about.
	// This procedure is done during options validation already,
	// which is why we don't check for errors here.
	moduleDirAbs, _ := filepath.Abs(moduleDir)
	mainFileAbs, _ := filepath.Abs(filepath.Join(moduleDirAbs, mainFile))

	// Construct path to mainFile relative to moduleDir
	mainFileRel := strings.TrimPrefix(mainFileAbs, moduleDirAbs)
	mainFileRel = strings.TrimPrefix(mainFileRel, "/")

	// The application name is the name of the directory that contains
	// the main file. There may be cases where this is not true.
	// We could add a -name flag to override this in the future.
	var applicationName string

	if mainDir, _ := filepath.Split(mainFileRel); mainDir != "" {
		mainDir = strings.TrimSuffix(mainDir, "/")
		applicationName = filepath.Base(mainDir)

		oldPURL := bom.Metadata.Component.PackageURL
		newPURL := oldPURL + "#" + mainDir

		log.Debug().
			Str("old", oldPURL).
			Str("new", newPURL).
			Msg("updating purl of main component")

		// Update PURL of main component
		bom.Metadata.Component.BOMRef = newPURL
		bom.Metadata.Component.PackageURL = newPURL

		// Update PURL in dependency graph
		for i, dep := range *bom.Dependencies {
			if dep.Ref == oldPURL {
				(*bom.Dependencies)[i].Ref = newPURL
				break
			}
		}
	} else {
		applicationName = filepath.Base(moduleDirAbs)
	}

	applicationNameProperty := sbom.NewProperty("application:name", applicationName)
	if bom.Metadata.Component.Properties == nil {
		bom.Metadata.Component.Properties = &[]cdx.Property{applicationNameProperty}
	} else {
		*bom.Metadata.Component.Properties = append([]cdx.Property{applicationNameProperty}, *bom.Metadata.Component.Properties...)
	}
}
