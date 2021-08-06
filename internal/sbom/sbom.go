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

package sbom

import (
	"crypto/md5"  // #nosec G501
	"crypto/sha1" // #nosec G505
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/sha3"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	modconv "github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/module"
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
	mainComponent, err := modconv.ToComponent(mainModule,
		modconv.WithComponentType(options.ComponentType),
		modconv.WithLicenses(),
		modconv.WithScope(""), // Main component can't have a scope
	)
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}

	components, err := modconv.ToComponents(modules,
		withModuleHashes(), modconv.WithLicenses(), modconv.WithTestScope(cdx.ScopeOptional))
	if err != nil {
		return nil, fmt.Errorf("failed to convert modules: %w", err)
	}

	log.Println("building dependency graph")
	dependencyGraph := BuildDependencyGraph(append(modules, mainModule))

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
		tool, err := BuildToolMetadata()
		if err != nil {
			return nil, fmt.Errorf("failed to build tool metadata: %w", err)
		}

		bom.Metadata.Timestamp = time.Now().Format(time.RFC3339)
		bom.Metadata.Tools = &[]cdx.Tool{*tool}
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if options.IncludeStdLib {
		log.Println("gathering info about standard library")
		stdComponent, err := BuildStdComponent()
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

func withModuleHashes() modconv.Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if m.Main || m.Vendored {
			return nil
		}

		h1, err := m.Hash()
		if err != nil {
			return fmt.Errorf("failed to calculate h1 hash: %w", err)
		}

		h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
		if err != nil {
			return fmt.Errorf("failed to base64 decode h1 hash: %w", err)
		}

		c.Hashes = &[]cdx.Hash{
			{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
		}

		return nil
	}
}

func BuildStdComponent() (*cdx.Component, error) {
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

func BuildDependencyGraph(modules []gomod.Module) []cdx.Dependency {
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

func BuildToolMetadata() (*cdx.Tool, error) {
	toolExePath, err := os.Executable()
	if err != nil {
		return nil, err
	}

	toolHashes, err := CalculateFileHashes(toolExePath,
		cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA512)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tool hashes: %w", err)
	}

	return &cdx.Tool{
		Vendor:  version.Author,
		Name:    version.Name,
		Version: version.Version,
		Hashes:  &toolHashes,
	}, nil
}

func CalculateFileHashes(filePath string, algos ...cdx.HashAlgorithm) ([]cdx.Hash, error) {
	if len(algos) == 0 {
		return make([]cdx.Hash, 0), nil
	}

	hashMap := make(map[cdx.HashAlgorithm]hash.Hash)
	hashWriters := make([]io.Writer, 0)

	for _, algo := range algos {
		var hashWriter hash.Hash

		switch algo { //exhaustive:ignore
		case cdx.HashAlgoMD5:
			hashWriter = md5.New() // #nosec G401
		case cdx.HashAlgoSHA1:
			hashWriter = sha1.New() // #nosec G401
		case cdx.HashAlgoSHA256:
			hashWriter = sha256.New()
		case cdx.HashAlgoSHA384:
			hashWriter = sha512.New384()
		case cdx.HashAlgoSHA512:
			hashWriter = sha512.New()
		case cdx.HashAlgoSHA3_256:
			hashWriter = sha3.New256()
		case cdx.HashAlgoSHA3_512:
			hashWriter = sha3.New512()
		default:
			return nil, fmt.Errorf("unsupported hash algorithm: %s", algo)
		}

		hashWriters = append(hashWriters, hashWriter)
		hashMap[algo] = hashWriter
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	multiWriter := io.MultiWriter(hashWriters...)
	if _, err = io.Copy(multiWriter, file); err != nil {
		return nil, err
	}
	file.Close()

	cdxHashes := make([]cdx.Hash, 0, len(hashMap))
	for _, algo := range algos { // Don't iterate over hashMap, as it doesn't retain order
		cdxHashes = append(cdxHashes, cdx.Hash{
			Algorithm: algo,
			Value:     fmt.Sprintf("%x", hashMap[algo].Sum(nil)),
		})
	}

	return cdxHashes, nil
}
