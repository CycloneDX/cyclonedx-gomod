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
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	modConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
	pkgConv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/pkg"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/generate"
	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
)

type generator struct {
	logger zerolog.Logger

	includeFiles    bool
	includePackages bool
	includePaths    bool
	includeStdlib   bool
	licenseDetector licensedetect.Detector
	mainDir         string
	moduleDir       string
	shortPURLs      bool
}

func NewGenerator(moduleDir string, opts ...Option) (generate.Generator, error) {
	g := generator{
		logger:    log.Logger,
		moduleDir: moduleDir,
	}

	var err error
	for _, opt := range opts {
		if err = opt(&g); err != nil {
			return nil, err
		}
	}

	return &g, nil
}

// Generate implements the generate.Generator interface.
func (g generator) Generate() (*cdx.BOM, error) {
	modules, err := gomod.LoadModulesFromPackages(g.logger, g.moduleDir, g.mainDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load modules: %w", err)
	}

	moduleDirAbs, err := filepath.Abs(g.moduleDir)
	if err != nil {
		return nil, fmt.Errorf("failed to make moduleDir absolute: %w", err)
	}

	mainDirAbs := filepath.Join(moduleDirAbs, g.mainDir)
	var appPackageIndex int
	appModuleIndex := slices.IndexFunc(modules, func(mod gomod.Module) bool {
		appPackageIndex = slices.IndexFunc(mod.Packages, func(pkg gomod.Package) bool {
			return pkg.Name == "main" && filepath.Clean(pkg.Dir) == mainDirAbs
		})
		return appPackageIndex >= 0
	})
	if appModuleIndex < 0 {
		return nil, fmt.Errorf("failed to find application module")
	}

	for i, module := range modules {
		if module.Path == gomod.StdlibModulePath {
			if g.includeStdlib {
				modules[appModuleIndex].Dependencies = append(modules[appModuleIndex].Dependencies, &modules[i])
				break
			} else {
				modules = append(modules[:i], modules[i+1:]...)
				if i < appModuleIndex {
					appModuleIndex--
				}
				break
			}
		}
	}

	// Dependencies need to be applied prior to determining the main
	// module's version, because `go mod graph` omits that version.
	err = gomod.ApplyModuleGraph(g.logger, g.moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to apply module graph: %w", err)
	}

	// Only the main module is missing its version in `go list` output.
	if modules[appModuleIndex].Main {
		modules[appModuleIndex].Version, err = gomod.GetModuleVersion(g.logger, modules[appModuleIndex].Dir)
		if err != nil {
			return nil, fmt.Errorf("failed to determine version of main module: %w", err)
		}
	}

	mainComponent, err := modConv.ToComponent(g.logger, modules[appModuleIndex],
		modConv.WithComponentType(cdx.ComponentTypeApplication),
		modConv.WithLicenses(g.licenseDetector),
		modConv.WithShortPURL(g.shortPURLs),
		modConv.WithPackages(g.includePackages,
			pkgConv.WithFiles(g.includeFiles, g.includePaths),
			pkgConv.WithShortPURL(g.shortPURLs)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}

	buildProperties, err := g.createBuildProperties()
	if err != nil {
		return nil, err
	}
	if mainComponent.Properties == nil {
		mainComponent.Properties = &buildProperties
	} else {
		*mainComponent.Properties = append(*mainComponent.Properties, buildProperties...)
	}

	components, err := modConv.ToComponents(g.logger, modules,
		modConv.WithLicenses(g.licenseDetector),
		modConv.WithModuleHashes(),
		modConv.WithShortPURL(g.shortPURLs),
		modConv.WithPackages(g.includePackages,
			pkgConv.WithFiles(g.includeFiles, g.includePaths),
			pkgConv.WithShortPURL(g.shortPURLs)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert modules: %w", err)
	}
	components = append(components[0:appModuleIndex], components[appModuleIndex+1:]...)

	dependencies := sbom.BuildDependencyGraph(modules)

	bom := cdx.NewBOM()
	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	bom.Components = &components
	bom.Dependencies = &dependencies

	g.includeAppPathInMainComponentPURL(bom, modules[appModuleIndex].Path, modules[appModuleIndex].Packages[appPackageIndex].ImportPath)

	return bom, nil
}

var buildEnv = []string{
	"CGO_ENABLED",
	"GOARCH",
	"GOOS",
	"GOVERSION",
}

func (g generator) createBuildProperties() (properties []cdx.Property, err error) {
	env, err := gocmd.GetEnv(g.logger)
	if err != nil {
		return nil, err
	}

	for _, buildEnvKey := range buildEnv {
		buildEnvVal, ok := env[buildEnvKey]
		if !ok {
			g.logger.Warn().
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

// includeAppPathInMainComponentPURL determines the path to the application
// (path to mainDir) relative to the module that owns the main package (moduleDir).
// If the application path is not equal to moduleDir, it is added to the main component's
// package URL as sub path, unless short PURLs are enabled. For example:
//
// + moduleDir 		<- application name
// |-+ main.go
//
// + moduleDir
// |-+ cmd
// ..|-+ mainDir	<- application name
// ....|-+ main.go
//
// The package URLs for the above examples would look like this:
//  1. pkg:golang/.../module@version         		(untouched)
//  2. pkg:golang/.../module@version#cmd/mainDir 	(with sub path)
//
// The selected main package may belong to a vendored tool rather than
// the module providing the dependency graph. The sub path is therefore derived
// from import paths relative to the owning module: selecting
// vendor/example.com/module/cmd/app yields #cmd/app within example.com/module.
//
// If the package URL is updated, the BOM reference is as well.
// All places within the BOM that reference the main component will be updated accordingly.
// This includes both dependencies[].ref and incoming dependsOn references.
func (g generator) includeAppPathInMainComponentPURL(bom *cdx.BOM, modulePath, packagePath string) {
	if subpath, ok := strings.CutPrefix(packagePath, modulePath+"/"); ok && !g.shortPURLs {
		oldPURL := bom.Metadata.Component.PackageURL
		newPURL := oldPURL + "#" + subpath

		oldBOMRef := bom.Metadata.Component.BOMRef
		newBOMRef := oldBOMRef + "#" + subpath

		g.logger.Debug().
			Str("oldpurl", oldPURL).
			Str("newpurl", newPURL).
			Str("oldbomref", oldBOMRef).
			Str("newbomref", newBOMRef).
			Msg("updating purl of main component")

		// Update BOMRef and PURL of main component
		bom.Metadata.Component.BOMRef = newBOMRef
		bom.Metadata.Component.PackageURL = newPURL

		// Update both outgoing and incoming refs in the dependency graph (without GOOS and GOARCH).
		for i, dep := range *bom.Dependencies {
			if dep.Ref == oldBOMRef {
				(*bom.Dependencies)[i].Ref = newBOMRef
			}
			if dep.Dependencies != nil {
				for j, ref := range *dep.Dependencies {
					if ref == oldBOMRef {
						(*dep.Dependencies)[j] = newBOMRef
					}
				}
			}
		}
	}
}
