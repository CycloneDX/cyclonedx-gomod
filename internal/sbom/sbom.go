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

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/license"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/google/uuid"
)

type GenerateOptions struct {
	ComponentType   cdx.ComponentType
	IncludeStdLib   bool
	NoSerialNumber  bool
	NoVersionPrefix bool
	SerialNumber    *uuid.UUID
}

func Generate(modulePath string, options GenerateOptions) (*cdx.BOM, error) {
	modules, err := gomod.GetModules(modulePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get modules: %w", err)
	}

	for i := range modules {
		modules[i].Version = strings.TrimSuffix(modules[i].Version, "+incompatible")

		if options.NoVersionPrefix {
			modules[i].Version = strings.TrimPrefix(modules[i].Version, "v")
		}
	}

	mainModule := modules[0]
	modules = modules[1:]

	if mainModule.Version, err = gomod.GetModuleVersion(mainModule.Dir); err != nil {
		log.Printf("failed to get version of main module: %v\n", err)
	}
	if mainModule.Version != "" && options.NoVersionPrefix {
		mainModule.Version = strings.TrimPrefix(mainModule.Version, "v")
	}

	mainComponent, err := convertToComponent(mainModule)
	if err != nil {
		return nil, fmt.Errorf("failed to convert main module: %w", err)
	}
	mainComponent.Scope = "" // Main component can't have a scope
	mainComponent.Type = options.ComponentType

	component := new(cdx.Component)
	components := make([]cdx.Component, len(modules))
	for i, module := range modules {
		component, err = convertToComponent(module)
		if err != nil {
			return nil, fmt.Errorf("failed to convert module %s: %w", module.Coordinates(), err)
		}
		components[i] = *component
	}

	dependencyGraph := buildDependencyGraph(append(modules, mainModule))

	toolHashes, err := calculateToolHashes()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate tool hashes: %w", err)
	}

	bom := cdx.NewBOM()
	if !options.NoSerialNumber {
		if options.SerialNumber == nil {
			bom.SerialNumber = uuid.New().URN()
		} else {
			bom.SerialNumber = options.SerialNumber.URN()
		}
	}

	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: &[]cdx.Tool{
			{
				Vendor:  version.Author,
				Name:    version.Name,
				Version: version.Version,
				Hashes:  &toolHashes,
			},
		},
		Component: mainComponent,
	}

	bom.Components = &components
	bom.Dependencies = &dependencyGraph

	if options.IncludeStdLib {
		stdComponent, err := buildStdComponent()
		if err != nil {
			return nil, fmt.Errorf("failed to build std component: %w", err)
		}

		*bom.Components = append(*bom.Components, *stdComponent)

		// Add std to dependency graph
		stdDependency := cdx.Dependency{Ref: stdComponent.BOMRef}
		*bom.Dependencies = append(*bom.Dependencies, stdDependency)

		// Add std as dependency of main module
		for _, dependency := range *bom.Dependencies {
			if dependency.Ref == mainComponent.BOMRef {
				*dependency.Dependencies = append(*dependency.Dependencies, stdDependency)
				break
			}
		}
	}

	return bom, nil
}

func convertToComponent(module gomod.Module) (*cdx.Component, error) {
	if module.Replace != nil {
		return convertToComponent(*module.Replace)
	}

	component := cdx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cdx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version,
		Scope:      cdx.ScopeRequired,
		PackageURL: module.PackageURL(),
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

	if !module.Main {
		resolvedLicense, err := license.Resolve(module)
		if err == nil {
			component.Licenses = &[]cdx.LicenseChoice{
				{
					License: &cdx.License{
						ID: resolvedLicense,
					},
				},
			}
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
	// See https://labix.org/gopkg.in
	// gopkg.in/pkg.v3		→ github.com/go-pkg/pkg	(branch/tag v3, v3.N, or v3.N.M)
	// gopkg.in/user/pkg.v3	→ github.com/user/pkg  	(branch/tag v3, v3.N, or v3.N.M)
	goPkgInRegex1 = regexp.MustCompile("^gopkg\\.in/([^/]+)/([^.]+)\\..*$") // With user segment
	goPkgInRegex2 = regexp.MustCompile("^gopkg\\.in/([^.]+)\\..*$")         // Without user segment
)

func resolveVcsURL(module gomod.Module) string {
	if util.StartsWith(module.Path, "github.com/") {
		return "https://" + module.Path
	} else if goPkgInRegex1.MatchString(module.Path) {
		return "https://" + goPkgInRegex1.ReplaceAllString(module.Path, "github.com/$1/$2")
	} else if goPkgInRegex2.MatchString(module.Path) {
		return "https://" + goPkgInRegex2.ReplaceAllString(module.Path, "github.com/go-$1/$1")
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
