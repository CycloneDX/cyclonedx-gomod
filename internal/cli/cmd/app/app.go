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
	"fmt"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/pkg/sbom/convert/module"
	pkgConv "github.com/CycloneDX/cyclonedx-gomod/pkg/sbom/convert/pkg"
	"os"
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/gomod"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod app", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "app",
		ShortHelp:  "Generate SBOMs for applications",
		ShortUsage: "cyclonedx-gomod app [FLAGS...] [MODULE_PATH]",
		LongHelp: `Generate SBOMs for applications.

In order to produce accurate SBOMs, build constraints must be configured
via environment variables. These build constraints should mimic the ones passed
to the "go build" command for the application.

Environment variables that act as build constraints are:
  - GOARCH       The target architecture (386, amd64, etc.)
  - GOOS         The target operating system (linux, windows, etc.)
  - CGO_ENABLED  Whether or not CGO is enabled
  - GOFLAGS      Flags that are passed to the Go command (e.g. build tags)

A complete overview of all environment variables can be found here:
  https://pkg.go.dev/cmd/go#hdr-Environment_variables

Applicable build constraints are included as properties of the main component.

Because build constraints influence Go's module selection, an SBOM should be generated
for each target in the build matrix.

The -main flag should be used to specify the path to the application's main package.
It must point to a directory within MODULE_PATH. If not set, MODULE_PATH is assumed.

In order to not only include modules, but also the packages within them,
the -packages flag can be used. Packages are represented as subcomponents of modules.

By passing -files, all files that would be included in a binary will be attached
as subcomponents of their respective package. File versions follow the v0.0.0-SHORTHASH pattern, 
where SHORTHASH is the first 12 characters of the file's SHA1 hash.
Because files are subcomponents of packages, -files can only be used in conjunction with -packages.

Examples:
  $ GOARCH=arm64 GOOS=linux GOFLAGS="-tags=foo,bar" cyclonedx-gomod app -output linux-arm64.bom.xml
  $ cyclonedx-gomod app -json -output acme-app.bom.json -files -licenses -main cmd/acme-app /usr/src/acme-module`,
		FlagSet: fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments (expected 1, got %d)", len(args))
			}
			if len(args) == 0 {
				options.ModuleDir = "."
			} else {
				options.ModuleDir = args[0]
			}

			options.LogOptions.ConfigureLogger()

			return Exec(options)
		},
	}
}

func Exec(options Options) error {
	err := options.Validate()
	if err != nil {
		return err
	}

	modules, err := gomod.LoadModulesFromPackages(options.ModuleDir, options.Main)
	if err != nil {
		return fmt.Errorf("failed to load modules: %w", err)
	}

	for i, module := range modules {
		if module.Path == gomod.StdlibModulePath {
			if options.IncludeStd {
				modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
				break
			} else {
				modules = append(modules[:i], modules[i+1:]...)
				break
			}
		}
	}

	// Dependencies need to be applied prior to determining the main
	// module's version, because `go mod graph` omits that version.
	err = gomod.ApplyModuleGraph(options.ModuleDir, modules)
	if err != nil {
		return fmt.Errorf("failed to apply module graph: %w", err)
	}

	// Determine version of main module
	modules[0].Version, err = gomod.GetModuleVersion(modules[0].Dir)
	if err != nil {
		return fmt.Errorf("failed to determine version of main module: %w", err)
	}

	// Convert main module
	mainComponent, err := modConv.ToComponent(modules[0],
		modConv.WithComponentType(cdx.ComponentTypeApplication),
		modConv.WithLicenses(options.ResolveLicenses),
		modConv.WithPackages(options.IncludePackages,
			pkgConv.WithFiles(options.IncludeFiles)),
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	// Build properties (e.g. the Go version) depend on the environment
	// and are thus only included when the SBOM doesn't have to be reproducible.
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

	// Convert the other modules
	components, err := modConv.ToComponents(modules[1:],
		modConv.WithLicenses(options.ResolveLicenses),
		modConv.WithModuleHashes(),
		modConv.WithPackages(options.IncludePackages,
			pkgConv.WithFiles(options.IncludeFiles)),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	bom := cdx.NewBOM()
	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}

	// Assemble metadata
	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	err = cliUtil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to add common metadata: %w", err)
	}

	bom.Components = &components
	dependencies := sbom.BuildDependencyGraph(modules)
	bom.Dependencies = &dependencies

	enrichWithApplicationDetails(bom, options.ModuleDir, options.Main)

	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
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
func enrichWithApplicationDetails(bom *cdx.BOM, moduleDir, mainPkgDir string) {
	// Resolve absolute paths to moduleDir and mainPkgDir.
	// Both may contain traversals or similar elements we don't care about.
	// This procedure is done during options validation already,
	// which is why we don't check for errors here.
	moduleDirAbs, _ := filepath.Abs(moduleDir)
	mainPkgDirAbs, _ := filepath.Abs(filepath.Join(moduleDirAbs, mainPkgDir))

	// Construct path to mainPkgDir relative to moduleDir
	mainPkgDirRel := strings.TrimPrefix(mainPkgDirAbs, moduleDirAbs)
	mainPkgDirRel = strings.TrimPrefix(mainPkgDirRel, string(os.PathSeparator))

	if mainPkgDirRel != "" {
		mainPkgDirRel = strings.TrimSuffix(mainPkgDirRel, string(os.PathSeparator))

		oldPURL := bom.Metadata.Component.PackageURL
		newPURL := oldPURL + "#" + filepath.ToSlash(mainPkgDirRel)

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
	}
}
