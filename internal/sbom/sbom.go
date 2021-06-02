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
// Copyright (c) Niklas Düster. All Rights Reserved.

package sbom

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/license"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
)

type GenerateOptions struct {
	ComponentType   cdx.ComponentType
	IncludeStdLib   bool
	IncludeTest     bool
	NoSerialNumber  bool
	NoVersionPrefix bool
	Reproducible    bool
	ResolveLicenses bool
	SerialNumber    *uuid.UUID
}

func Generate(modulePath string, options GenerateOptions) (*cdx.BOM, error) {
	// Cheap trick to make Go download all required modules in the module graph
	// without modifying go.sum (as `go mod download` would do).
	log.Println("downloading modules")
	if err := gocmd.ModWhy(modulePath, []string{"github.com/CycloneDX/cyclonedx-go"}, io.Discard); err != nil {
		return nil, fmt.Errorf("downloading modules failed: %w", err)
	}

	log.Println("enumerating modules")
	modules, err := gomod.GetModules(modulePath, options.IncludeTest)
	if err != nil {
		return nil, fmt.Errorf("failed to enumerate modules: %w", err)
	}

	log.Println("normalizing module versions")
	for i := range modules {
		modules[i].Version = strings.TrimSuffix(modules[i].Version, "+incompatible")

		if options.NoVersionPrefix {
			modules[i].Version = strings.TrimPrefix(modules[i].Version, "v")
		}
	}

	mainModule := modules[0]
	modules = modules[1:]

	log.Println("determining version of main module")
	if mainModule.Version, err = gomod.GetModuleVersion(mainModule.Dir); err != nil {
		log.Printf("failed to get version of main module: %v\n", err)
	}
	if mainModule.Version != "" && options.NoVersionPrefix {
		mainModule.Version = strings.TrimPrefix(mainModule.Version, "v")
	}

	log.Printf("converting main module %s\n", mainModule.Coordinates())
	mainComponent, err := convertToComponent(mainModule, options.ResolveLicenses)
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}
	mainComponent.Scope = "" // Main component can't have a scope
	mainComponent.Type = options.ComponentType

	component := new(cdx.Component)
	components := make([]cdx.Component, len(modules))
	for i, module := range modules {
		component, err = convertToComponent(module, options.ResolveLicenses)
		if err != nil {
			return nil, fmt.Errorf("failed to convert module %s: %w", module.Coordinates(), err)
		}
		components[i] = *component
	}

	log.Println("building dependency graph")
	dependencyGraph := buildDependencyGraph(append(modules, mainModule))

	log.Println("calculating tool hashes")
	toolHashes := make([]cdx.Hash, 0)
	if !options.Reproducible {
		toolHashes, err = calculateToolHashes()
		if err != nil {
			return nil, fmt.Errorf("failed to calculate tool hashes: %w", err)
		}
	}

	log.Println("assembling sbom")
	bom := cdx.NewBOM()
	if !options.NoSerialNumber {
		if options.SerialNumber == nil {
			bom.SerialNumber = uuid.New().URN()
		} else {
			bom.SerialNumber = options.SerialNumber.URN()
		}
	}

	bom.Metadata = &cdx.Metadata{
		Component: mainComponent,
	}
	if !options.Reproducible {
		bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
		bom.Metadata.Tools = &[]cdx.Tool{
			{
				Vendor:  version.Author,
				Name:    version.Name,
				Version: version.Version,
				Hashes:  &toolHashes,
			},
		}
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if options.IncludeStdLib {
		log.Println("gathering info about standard library")
		stdComponent, err := buildStdComponent()
		if err != nil {
			return nil, fmt.Errorf("failed to build std component: %w", err)
		}

		log.Println("adding standard library to sbom")
		*bom.Components = append(*bom.Components, *stdComponent)

		// Add std to dependency graph
		stdDependency := cdx.Dependency{Ref: stdComponent.BOMRef}
		*bom.Dependencies = append(*bom.Dependencies, stdDependency)

		// Add std as dependency of main module
		for i, dependency := range *bom.Dependencies {
			if dependency.Ref == mainComponent.BOMRef {
				if dependency.Dependencies == nil {
					(*bom.Dependencies)[i].Dependencies = &[]cdx.Dependency{stdDependency}
				} else {
					*dependency.Dependencies = append(*dependency.Dependencies, stdDependency)
				}
				break
			}
		}
	}

	return bom, nil
}

func convertToComponent(module gomod.Module, resolveLicense bool) (*cdx.Component, error) {
	if module.Replace != nil {
		return convertToComponent(*module.Replace, resolveLicense)
	}

	log.Printf("converting module %s\n", module.Coordinates())

	component := cdx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cdx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version,
		PackageURL: module.PackageURL(),
	}

	if module.TestOnly {
		component.Scope = cdx.ScopeOptional
	} else {
		component.Scope = cdx.ScopeRequired
	}

	// We currently don't have an accurate way of hashing the main module, as it may contain
	// files that are .gitignore'd and thus not part of the hashes in Go's sumdb.
	//
	// Go's vendoring mechanism doesn't copy all files that make up a module to the vendor dir.
	// Hashing vendored modules thus won't result in the expected hash, probably causing more
	// confusion than anything else.
	//
	// TODO: Research how we can provide accurate hashes for main modules
	// TODO: Research how we can provide meaningful hashes for vendored modules
	if !module.Main && !module.Vendored {
		hashes, err := calculateModuleHashes(module)
		if err != nil {
			return nil, err
		}
		component.Hashes = &hashes
	}

	private, err := module.Private()
	if err != nil {
		// An error indicates a bad pattern, which must be fixed before we can proceed
		return nil, fmt.Errorf("failed to determine if module is private: %w", err)
	}

	if resolveLicense && !module.Main && !private {
		resolvedLicenses, err := license.Resolve(module)
		if err == nil {
			componentLicenses := make([]cdx.LicenseChoice, len(resolvedLicenses))
			for i := range resolvedLicenses {
				componentLicenses[i] = cdx.LicenseChoice{License: &resolvedLicenses[i]}
			}
			component.Licenses = &componentLicenses
		} else {
			log.Printf("failed to resolve license of %s: %v\n", module.Coordinates(), err)
		}
	}

	if vcsURL := resolveVcsURL(module); vcsURL != "" {
		component.ExternalReferences = &[]cdx.ExternalReference{
			{Type: cdx.ERTypeVCS, URL: vcsURL},
		}
	}

	return &component, nil
}

func calculateModuleHashes(module gomod.Module) ([]cdx.Hash, error) {
	h1, err := module.Hash()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate h1 hash: %w", err)
	}

	h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode h1 hash: %w", err)
	}

	return []cdx.Hash{
		{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
	}, nil
}

var (
	// By convention, modules with a major version equal to or above v2
	// have it as suffix in their module path.
	vcsUrlMajorVersionSuffixRegex = regexp.MustCompile(`(/v[\d]+)$`)

	// gopkg.in with user segment
	// Example: gopkg.in/user/pkg.v3 -> github.com/user/pkg
	vcsUrlGoPkgInRegexWithUser = regexp.MustCompile(`^gopkg\.in/([^/]+)/([^.]+)\..*$`)

	// gopkg.in without user segment
	// Example: gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
	vcsUrlGoPkgInRegexWithoutUser = regexp.MustCompile(`^gopkg\.in/([^.]+)\..*$`)
)

func resolveVcsURL(module gomod.Module) string {
	if strings.HasPrefix(module.Path, "github.com/") {
		return "https://" + vcsUrlMajorVersionSuffixRegex.ReplaceAllString(module.Path, "")
	} else if vcsUrlGoPkgInRegexWithUser.MatchString(module.Path) {
		return "https://" + vcsUrlGoPkgInRegexWithUser.ReplaceAllString(module.Path, "github.com/$1/$2")
	} else if vcsUrlGoPkgInRegexWithoutUser.MatchString(module.Path) {
		return "https://" + vcsUrlGoPkgInRegexWithoutUser.ReplaceAllString(module.Path, "github.com/go-$1/$1")
	}
	return ""
}

func buildStdComponent() (*cdx.Component, error) {
	goVersion, err := gocmd.GetVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to determine Go version: %w", err)
	}
	goVersion = strings.TrimPrefix(goVersion, "go")
	stdPURL := "pkg:golang/std@" + goVersion

	return &cdx.Component{
		BOMRef:      stdPURL,
		Type:        cdx.ComponentTypeLibrary,
		Name:        "std",
		Version:     goVersion,
		Description: "The Go standard library",
		Scope:       cdx.ScopeRequired,
		PackageURL:  stdPURL,
		ExternalReferences: &[]cdx.ExternalReference{
			{
				Type: cdx.ERTypeDocumentation,
				URL:  "https://golang.org/pkg/",
			},
			{
				Type: cdx.ERTypeVCS,
				URL:  "https://go.googlesource.com/go",
			},
			{
				Type: cdx.ERTypeWebsite,
				URL:  "https://golang.org/",
			},
		},
	}, nil
}

func buildDependencyGraph(modules []gomod.Module) []cdx.Dependency {
	depGraph := make([]cdx.Dependency, 0)

	for _, module := range modules {
		if module.Replace != nil {
			module = *module.Replace
		}
		cdxDependant := cdx.Dependency{Ref: module.PackageURL()}

		if module.Dependencies != nil {
			cdxDependencies := make([]cdx.Dependency, len(module.Dependencies))
			for i := range module.Dependencies {
				if module.Dependencies[i].Replace != nil {
					cdxDependencies[i] = cdx.Dependency{Ref: module.Dependencies[i].Replace.PackageURL()}
				} else {
					cdxDependencies[i] = cdx.Dependency{Ref: module.Dependencies[i].PackageURL()}
				}
			}
			if len(cdxDependencies) > 0 {
				cdxDependant.Dependencies = &cdxDependencies
			}
		}
		depGraph = append(depGraph, cdxDependant)
	}

	return depGraph
}

func calculateToolHashes() ([]cdx.Hash, error) {
	exePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	exeFile, err := os.Open(exePath)
	if err != nil {
		return nil, err
	}
	defer exeFile.Close()

	hashMD5 := md5.New()
	hashSHA1 := sha1.New()
	hashSHA256 := sha256.New()
	hashSHA512 := sha512.New()
	hashWriter := io.MultiWriter(hashMD5, hashSHA1, hashSHA256, hashSHA512)

	if _, err = io.Copy(hashWriter, exeFile); err != nil {
		return nil, err
	}

	return []cdx.Hash{
		{Algorithm: cdx.HashAlgoMD5, Value: fmt.Sprintf("%x", hashMD5.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA1, Value: fmt.Sprintf("%x", hashSHA1.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", hashSHA256.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA512, Value: fmt.Sprintf("%x", hashSHA512.Sum(nil))},
	}, nil
}
