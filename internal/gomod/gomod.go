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

package gomod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"github.com/rs/zerolog/log"
	"golang.org/x/mod/semver"
	"golang.org/x/mod/sumdb/dirhash"
)

// ErrNoGoModule indicates that a given path is not a valid Go module
var ErrNoGoModule = errors.New("not a Go module")

// See https://golang.org/ref/mod#go-list-m
type Module struct {
	Path    string  // module path
	Version string  // module version
	Replace *Module // replaced by this module
	Main    bool    // is this the main module?
	Dir     string  // directory holding files for this module, if any

	Dependencies []*Module `json:"-"` // modules this module depends on
	Local        bool      `json:"-"` // is this a local module?
	TestOnly     bool      `json:"-"` // is this module only required for tests?
	Vendored     bool      `json:"-"` // is this a vendored module?

	Files     []string `json:"-"` // files in this module
	TestFiles []string `json:"-"` // test files in this module
}

func (m Module) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}
	return m.Path + "@" + m.Version
}

func (m Module) Hash() (string, error) {
	h1, err := dirhash.HashDir(m.Dir, m.Coordinates(), dirhash.Hash1)
	if err != nil {
		return "", err
	}

	return h1, nil
}

func (m Module) PackageURL() string {
	return "pkg:golang/" + m.Coordinates()
}

func GetModules(mainModulePath string, includeTest bool) ([]Module, error) {
	if !util.IsGoModule(mainModulePath) {
		return nil, ErrNoGoModule
	}

	var (
		modules []Module
		err     error
	)

	// We're going to call the go command a few times and
	// we'll (re-)use this buffer to write its output to
	// and subsequently read from it.
	buf := new(bytes.Buffer)

	if !util.IsVendoring(mainModulePath) {
		if err = gocmd.ListModules(mainModulePath, buf); err != nil {
			return nil, fmt.Errorf("listing modules failed: %w", err)
		}

		modules, err = parseModules(buf)
		if err != nil {
			return nil, fmt.Errorf("parsing modules failed: %w", err)
		}
	} else {
		if err = gocmd.ListVendoredModules(mainModulePath, buf); err != nil {
			return nil, fmt.Errorf("listing vendored modules failed: %w", err)
		}

		modules, err = parseVendoredModules(mainModulePath, buf)
		if err != nil {
			return nil, fmt.Errorf("parsing vendored modules failed: %w", err)
		}

		// Main module is not included in vendored module list,
		// so we have to get it separately
		buf.Reset()
		if err = gocmd.GetModule(mainModulePath, buf); err != nil {
			return nil, fmt.Errorf("listing main module failed: %w", err)
		}

		mainModule := Module{}
		if err = json.NewDecoder(buf).Decode(&mainModule); err != nil {
			return nil, fmt.Errorf("parsing main module failed: %w", err)
		}

		modules = append(modules, mainModule)
	}

	modules, err = filterModules(mainModulePath, modules, includeTest)
	if err != nil {
		return nil, fmt.Errorf("filtering modules failed: %w", err)
	}

	SortModules(modules)

	// Replacements may point to local directories, in which case their .Path is
	// not the actual module's name, but the filepath as used in go.mod.
	for i := range modules {
		if modules[i].Replace == nil {
			continue
		}

		var localModulePath string
		if filepath.IsAbs(modules[i].Replace.Path) {
			localModulePath = modules[i].Replace.Path
		} else {
			localModulePath = filepath.Join(mainModulePath, modules[i].Replace.Path)
		}
		if !util.IsGoModule(localModulePath) {
			continue
		}

		if err = resolveLocalModule(localModulePath, modules[i].Replace); err != nil {
			return nil, fmt.Errorf("resolving local module %s failed: %w", modules[i].Replace.Coordinates(), err)
		}
	}

	buf.Reset()
	if err = gocmd.GetModuleGraph(mainModulePath, buf); err != nil {
		return nil, fmt.Errorf("listing module graph failed: %w", err)
	}

	if err = parseModuleGraph(buf, modules); err != nil {
		return nil, fmt.Errorf("parsing module graph failed: %w", err)
	}

	return modules, nil
}

// parseModules parses the output of `go list -json -m` into a Module slice.
func parseModules(reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)
	jsonDecoder := json.NewDecoder(reader)

	// Output is not a JSON array, so we have to parse one object after another
	for {
		var mod Module
		if err := jsonDecoder.Decode(&mod); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		modules = append(modules, mod)
	}
	return modules, nil
}

// parseVendoredModules parses the output of `go mod vendor -v` into a Module slice.
func parseVendoredModules(mainModulePath string, reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "# ") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, "# "))

		// Replacements may be specified as
		//   Path [Version] => Path [Version]
		arrowIndex := util.StringSliceIndex(fields, "=>")

		if arrowIndex == -1 {
			if len(fields) != 2 {
				return nil, fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
			}

			modules = append(modules, Module{
				Path:     fields[0],
				Version:  fields[1],
				Dir:      filepath.Join(mainModulePath, "vendor", fields[0]),
				Vendored: true,
			})
		} else {
			pathParent := fields[0]
			versionParent := ""
			if arrowIndex == 2 {
				versionParent = fields[1]
			}

			pathReplacement := fields[arrowIndex+1]
			versionReplacement := ""
			if len(fields) == arrowIndex+3 {
				versionReplacement = fields[arrowIndex+2]
			}

			modules = append(modules, Module{
				Path:    pathParent,
				Version: versionParent,
				Replace: &Module{
					Path:     pathReplacement,
					Version:  versionReplacement,
					Dir:      filepath.Join(mainModulePath, "vendor", pathParent), // Replacements are copied to their parents dir
					Vendored: true,
				},
			})
		}
	}

	return modules, nil
}

// parseModuleGraph parses the output of `go mod graph` and populates
// the .Dependencies field of a given Module slice. Dependencies are
// sorted by module path in ascending order.
//
// The Module slice is expected to contain only "effective" modules,
// with only a single version per module, as provided by `go list -m` or `go list -deps`.
func parseModuleGraph(reader io.Reader, modules []Module) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
		}

		// The module graph contains dependency relationships for multiple versions of a module.
		// When identifying the ACTUAL dependant, we search for it in strict mode (versions must match).
		dependant := findModule(modules, fields[0], true)
		if dependant == nil {
			continue
		}

		// The identified module may depend on an older version of its dependency.
		// Due to Go's minimal version selection, that version may not be present in
		// the effective modules slice. Hence, we search for the dependency in non-strict mode.
		dependency := findModule(modules, fields[1], false)
		if dependency == nil {
			log.Debug().
				Str("dependant", dependant.Coordinates()).
				Str("dependency", fields[1]).
				Msg("dependency not found")
			continue
		}

		if dependant.Dependencies == nil {
			dependant.Dependencies = []*Module{dependency}
		} else {
			dependant.Dependencies = append(dependant.Dependencies, dependency)
		}
	}

	for i := range modules {
		SortDependencies(modules[i].Dependencies)
	}

	return nil
}

// parseModWhy parses the output of `go mod why`,
// populating a map with module paths as keys and a list of packages as values.
func parseModWhy(reader io.Reader) map[string][]string {
	modPkgs := make(map[string][]string)
	currentModPath := ""

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "(main module does not need ") {
			continue
		}

		if strings.HasPrefix(line, "#") {
			currentModPath = strings.TrimSpace(strings.TrimPrefix(line, "#"))
			modPkgs[currentModPath] = make([]string, 0)
			continue
		}

		modPkgs[currentModPath] = append(modPkgs[currentModPath], line)
	}

	return modPkgs
}

func GetModulesFromBinary(binaryPath string) ([]Module, map[string]string, error) {
	buf := new(bytes.Buffer)
	if err := gocmd.GetModulesFromBinary(binaryPath, buf); err != nil {
		return nil, nil, err
	}

	modules, hashes := parseModulesFromBinary(buf)
	SortModules(modules)

	return modules, hashes, nil
}

func parseModulesFromBinary(reader io.Reader) ([]Module, map[string]string) {
	modules := make([]Module, 0)
	hashes := make(map[string]string)

	moduleIndex := 0
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Fields(line)
		switch fields[0] {
		case "mod": // Main module
			modules = append(modules, Module{
				Path:    fields[1],
				Version: fields[2],
				Main:    true,
			})
			moduleIndex += 1
		case "dep": // Depdendency module
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			modules = append(modules, module)
			if len(fields) == 4 {
				// Hash won't be available when the module is replaced
				hashes[module.Coordinates()] = fields[3]
			}
			moduleIndex += 1
		case "=>": // Replacement
			module := Module{
				Path:    fields[1],
				Version: fields[2],
			}
			modules[moduleIndex-1].Replace = &module
			hashes[module.Coordinates()] = fields[3]
		}
	}

	return modules, hashes
}

func findModule(modules []Module, coordinates string, strict bool) *Module {
	for i := range modules {
		if coordinates == modules[i].Coordinates() || (!strict && strings.HasPrefix(coordinates, modules[i].Path+"@")) {
			if modules[i].Replace != nil {
				return modules[i].Replace
			}
			return &modules[i]
		}
	}
	return nil
}

// filterModules queries `go mod why` with all provided modules to determine whether or not
// they're required by the main module. Modules required by the main module are returned in
// a new slice.
//
// Unless includeTest is true, test-only dependencies are not included in the returned slice.
// Test-only modules will have the TestOnly field set to true.
//
// Note that this method doesn't work when replacements have already been applied to the module slice.
// Consider a go.mod file containing the following lines:
//
// 		require golang.org/x/crypto v0.0.0-xxx-xxx
//		replace golang.org/x/crypto => github.com/ProtonMail/go-crypto v0.0.0-xxx-xxx
//
// Querying `go mod why -m` with `golang.org/x/crypto` yields the expected result, querying it with
// `github.com/ProtonMail/go-crypto` will always yield `(main module does not need github.com/ProtonMail/go-crypto)`.
// See:
//   - https://github.com/golang/go/issues/30720
//   - https://github.com/golang/go/issues/26904
func filterModules(mainModulePath string, modules []Module, includeTest bool) ([]Module, error) {
	buf := new(bytes.Buffer)
	filtered := make([]Module, 0)
	chunks := chunkModules(modules, 20)

	for _, chunk := range chunks {
		paths := make([]string, len(chunk))
		for i := range chunk {
			paths[i] = chunk[i].Path
		}

		if err := gocmd.ModWhy(mainModulePath, paths, buf); err != nil {
			return nil, err
		}

		for modPath, modPkgs := range parseModWhy(buf) {
			if len(modPkgs) == 0 {
				log.Debug().Str("module", modPath).Msg("filtering unneeded module")
				continue
			}

			// If the shortest package path contains test nodes, this is a test-only dependency.
			testOnly := false
			for _, pkg := range modPkgs {
				if strings.HasSuffix(pkg, ".test") {
					testOnly = true
					break
				}
			}
			if !includeTest && testOnly {
				log.Debug().Str("module", modPath).Msg("filtering test-only module")
				continue
			}

			for i := range chunk {
				if chunk[i].Path == modPath {
					mod := chunk[i]
					mod.TestOnly = testOnly
					filtered = append(filtered, mod)
				}
			}
		}

		buf.Reset()
	}

	return filtered, nil
}

// SortModules sorts a given Module slice ascendingly by path.
// Main modules take precedence, so that they will represent the first elements of the sorted slice.
// If the path of two modules are equal, they'll be compared by their semantic version instead.
func SortModules(modules []Module) {
	sort.Slice(modules, func(i, j int) bool {
		if modules[i].Main && !modules[j].Main {
			return true
		} else if !modules[i].Main && modules[j].Main {
			return false
		}

		if modules[i].Path == modules[j].Path {
			return semver.Compare(modules[i].Version, modules[j].Version) == -1
		}

		return modules[i].Path < modules[j].Path
	})
}

// SortDependencies sorts a given Module pointer slice ascendingly by path.
// If the path of two modules are equal, they'll be compared by their semantic version instead.
func SortDependencies(dependencies []*Module) {
	sort.Slice(dependencies, func(i, j int) bool {
		if dependencies[i].Path == dependencies[j].Path {
			return semver.Compare(dependencies[i].Version, dependencies[j].Version) == -1
		}

		return dependencies[i].Path < dependencies[j].Path
	})
}

func resolveLocalModule(localModulePath string, module *Module) error {
	if util.IsGoModule(module.Dir) && strings.HasPrefix(module.Dir, util.GetModuleCacheDir()) {
		// Module is in module cache
		return nil
	} else if !util.IsGoModule(localModulePath) {
		return ErrNoGoModule
	}

	buf := new(bytes.Buffer)
	if err := gocmd.GetModule(localModulePath, buf); err != nil {
		return err
	}
	localModule := new(Module)
	if err := json.NewDecoder(buf).Decode(localModule); err != nil {
		return err
	}

	module.Path = localModule.Path
	module.Local = true

	// Try to resolve the version. Only works when module.Dir is a Git repo.
	if module.Version == "" {
		version, err := GetModuleVersion(module.Dir)
		if err == nil {
			module.Version = version
		} else {
			// We don't fail with an error here, because our possibilities are limited.
			// module.Dir may be a Mercurial repo or just a normal directory, in which case we
			// cannot detect versions reliably right now.
			log.Warn().Err(err).Str("module", module.Path).Msg("failed to resolve version of local module")
		}
	}

	return nil
}

func chunkModules(modules []Module, chunkSize int) [][]Module {
	var chunks [][]Module

	for i := 0; i < len(modules); i += chunkSize {
		j := i + chunkSize

		if j > len(modules) {
			j = len(modules)
		}

		chunks = append(chunks, modules[i:j])
	}

	return chunks
}
