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
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/module"

	cdx "github.com/CycloneDX/cyclonedx-go"
	cliUtil "github.com/CycloneDX/cyclonedx-gomod/internal/cli/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/rs/zerolog/log"
)

func New() *ffcli.Command {
	fs := flag.NewFlagSet("cyclonedx-gomod bin", flag.ExitOnError)

	var options Options
	options.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "bin",
		ShortHelp:  "Generate SBOMs for binaries",
		ShortUsage: "cyclonedx-gomod bin [FLAGS...] BINARY_PATH",
		LongHelp: `Generate SBOMs for binaries.

Although the binary is never executed by cyclonedx-gomod, it must be executable.
This is a requirement by the "go version -m" command that is used to provide this functionality.

When license detection is enabled, all modules (including the main module) 
will be downloaded to the module cache using "go mod download".
For the download of the main module to work, its version has to be provided
via the -version flag.

Please note that data embedded in binaries shouldn't be trusted,
unless there's solid evidence that the binaries haven't been modified
since they've been built.

Example:
  $ cyclonedx-gomod bin -json -output acme-app-v1.0.0.bom.json -version v1.0.0 ./acme-app`,
		FlagSet: fs,
		Exec: func(_ context.Context, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("too many arguments (expected 1, got %d)", len(args))
			}
			if len(args) == 1 {
				options.BinaryPath = args[0]
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

	bi, err := gomod.LoadBuildInfo(options.BinaryPath)
	if err != nil {
		return fmt.Errorf("failed to load build info: %w", err)
	} else if bi.Main == nil {
		return fmt.Errorf("failed to parse any modules from %s", options.BinaryPath)
	}

	modules := append([]gomod.Module{*bi.Main}, bi.Deps...)

	if options.IncludeStd {
		modules = append(modules, gomod.Module{
			Path:    gomod.StdlibModulePath,
			Version: bi.GoVersion,
		})
	}

	if options.Version != "" {
		modules[0].Version = options.Version
	} else if modules[0].Version == "(devel)" && len(bi.Settings) > 0 {
		log.Debug().Msg("building pseudo version from buildinfo")
		modules[0].Version, err = buildPseudoVersion(bi)
		if err != nil {
			log.Warn().Err(err).Msg("failed to build pseudo version from buildinfo")
		}
	}

	// If we want to resolve licenses, we have to download the modules first
	if options.ResolveLicenses {
		err = downloadModules(modules)
		if err != nil {
			return fmt.Errorf("failed to download modules: %w", err)
		}
	}

	// Make all modules a direct dependency of the main module
	for i := 1; i < len(modules); i++ {
		modules[0].Dependencies = append(modules[0].Dependencies, &modules[i])
	}

	// Convert main module
	mainComponent, err := modConv.ToComponent(modules[0],
		modConv.WithComponentType(cdx.ComponentTypeApplication),
		modConv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert main module: %w", err)
	}

	// Convert the other modules
	components, err := modConv.ToComponents(modules[1:],
		modConv.WithLicenses(options.ResolveLicenses),
	)
	if err != nil {
		return fmt.Errorf("failed to convert modules: %w", err)
	}

	binaryProperties, err := buildBinaryProperties(options.BinaryPath, bi)
	if err != nil {
		return fmt.Errorf("failed to create binary properties")
	}

	bom := cdx.NewBOM()

	err = cliUtil.SetSerialNumber(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to set serial number: %w", err)
	}

	bom.Metadata = &cdx.Metadata{
		Component:  mainComponent,
		Properties: &binaryProperties,
	}
	err = cliUtil.AddCommonMetadata(bom, options.SBOMOptions)
	if err != nil {
		return fmt.Errorf("failed to add common metadata")
	}

	bom.Components = &components
	dependencyGraph := sbom.BuildDependencyGraph(modules)
	bom.Dependencies = &dependencyGraph

	if bi.Path != bi.Main.Path && strings.HasPrefix(bi.Path, bi.Main.Path) {
		subpath := strings.TrimPrefix(bi.Path, bi.Main.Path)
		subpath = strings.TrimPrefix(subpath, "/")

		oldPURL := bom.Metadata.Component.PackageURL
		newPURL := oldPURL + "#" + subpath

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

		// Because buildCompositions works on components and not modules,
		// the updated PURL will be reflected in there without further ado.
	}

	bom.Compositions = buildCompositions(*mainComponent, components)

	if options.AssertLicenses {
		sbom.AssertLicenses(bom)
	}

	return cliUtil.WriteBOM(bom, options.OutputOptions)
}

// buildPseudoVersion builds a pseudo version for the main module.
// Requires that the binary was built with Go 1.18+ and the build
// settings include VCS information.
//
// Because major version and previous version are not known,
// this operation will always produce a v0.0.0-TIME-REF version.
func buildPseudoVersion(bi *gomod.BuildInfo) (string, error) {
	vcsRev, ok := bi.Settings["vcs.revision"]
	if !ok {
		return "", fmt.Errorf("no vcs.revision buildinfo")
	}
	if len(vcsRev) > 12 {
		vcsRev = vcsRev[:12]
	}
	vcsTimeStr, ok := bi.Settings["vcs.time"]
	if !ok {
		return "", fmt.Errorf("no vcs.time buildinfo")
	}
	vcsTime, err := time.Parse(time.RFC3339, vcsTimeStr)
	if err != nil {
		return "", err
	}

	return module.PseudoVersion("", "", vcsTime, vcsRev), nil
}

func buildBinaryProperties(binaryPath string, bi *gomod.BuildInfo) ([]cdx.Property, error) {
	properties := []cdx.Property{
		sbom.NewProperty("binary:name", filepath.Base(binaryPath)),
		sbom.NewProperty("build:env:GOVERSION", bi.GoVersion),
	}

	if len(bi.Settings) > 0 {
		if cgo, ok := bi.Settings["CGO_ENABLED"]; ok {
			properties = append(properties, sbom.NewProperty("build:env:CGO_ENABLED", cgo))
		}
		if goarch, ok := bi.Settings["GOARCH"]; ok {
			properties = append(properties, sbom.NewProperty("build:env:GOARCH", goarch))
		}
		if goos, ok := bi.Settings["GOOS"]; ok {
			properties = append(properties, sbom.NewProperty("build:env:GOOS", goos))
		}
		if compiler, ok := bi.Settings["-compiler"]; ok {
			properties = append(properties, sbom.NewProperty("build:compiler", compiler))
		}
		if tags, ok := bi.Settings["-tags"]; ok {
			for _, tag := range strings.Split(tags, ",") {
				properties = append(properties, sbom.NewProperty("build:tag", tag))
			}
		}
		if vcs, ok := bi.Settings["vcs"]; ok {
			properties = append(properties, sbom.NewProperty("build:vcs", vcs))
		}
		if vcsRev, ok := bi.Settings["vcs.revision"]; ok {
			properties = append(properties, sbom.NewProperty("build:vcs:revision", vcsRev))
		}
		if vcsTime, ok := bi.Settings["vcs.time"]; ok {
			properties = append(properties, sbom.NewProperty("build:vcs:time", vcsTime))
		}
		if vcsModified, ok := bi.Settings["vcs.modified"]; ok {
			properties = append(properties, sbom.NewProperty("build:vcs:modified", vcsModified))
		}
	}

	binaryHashes, err := sbom.CalculateFileHashes(binaryPath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate binary hashes: %w", err)
	}
	for _, hash := range binaryHashes {
		properties = append(properties, sbom.NewProperty(fmt.Sprintf("binary:hash:%s", hash.Algorithm), hash.Value))
	}

	sbom.SortProperties(properties)

	return properties, nil
}

func buildCompositions(mainComponent cdx.Component, components []cdx.Component) *[]cdx.Composition {
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

	return &compositions
}

func downloadModules(modules []gomod.Module) error {
	modulesToDownload := make([]gomod.Module, 0)
	for i := range modules {
		if modules[i].Path == gomod.StdlibModulePath {
			continue // We can't download the stdlib
		}

		// When modules are replaced, only download the replacement.
		if modules[i].Replace != nil {
			modulesToDownload = append(modulesToDownload, *modules[i].Replace)
		} else {
			modulesToDownload = append(modulesToDownload, modules[i])
		}
	}

	downloads, err := gomod.Download(modulesToDownload)
	if err != nil {
		return err
	}

	for i, download := range downloads {
		if download.Error != "" {
			log.Warn().
				Str("module", download.Coordinates()).
				Str("reason", download.Error).
				Msg("module download failed")
			continue
		}

		mm := matchModule(modules, download.Coordinates())
		if mm == nil {
			log.Warn().
				Str("module", download.Coordinates()).
				Msg("downloaded module not found")
			continue
		}

		// Check that the hash of the downloaded module matches
		// the one found in the binary. We want to report the version
		// for the *exact* module version or nothing at all.
		if mm.Sum != "" && mm.Sum != download.Sum {
			log.Warn().
				Str("binaryHash", mm.Sum).
				Str("downloadHash", download.Sum).
				Str("module", download.Coordinates()).
				Msg("module hash mismatch")
			continue
		}

		log.Debug().
			Str("module", download.Coordinates()).
			Msg("module downloaded")

		mm.Dir = downloads[i].Dir
	}

	return nil
}

func matchModule(modules []gomod.Module, coordinates string) *gomod.Module {
	for i, m := range modules {
		if m.Replace != nil && coordinates == m.Replace.Coordinates() {
			return modules[i].Replace
		}
		if coordinates == m.Coordinates() {
			return &modules[i]
		}
	}

	return nil
}
