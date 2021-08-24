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

// IsModule determines whether dir is a Go module.
func IsModule(dir string) bool {
	return util.FileExists(filepath.Join(dir, "go.mod"))
}

// ErrNoModule indicates that a given path is not a valid Go module
var ErrNoModule = errors.New("not a go module")

func GetModules(moduleDir string, includeTest bool) ([]Module, error) {
	if !util.IsGoModule(moduleDir) {
		return nil, ErrNoModule
	}

	buf := new(bytes.Buffer)

	err := gocmd.ListModules(moduleDir, buf)
	if err != nil {
		return nil, fmt.Errorf("listing modules failed: %w", err)
	}

	modules, err := parseModules(buf)
	if err != nil {
		return nil, fmt.Errorf("parsing modules failed: %w", err)
	}

	modules, err = FilterModules(moduleDir, modules, includeTest)
	if err != nil {
		return nil, fmt.Errorf("filtering modules failed: %w", err)
	}

	err = resolveLocalModules(moduleDir, modules)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve local modules")
	}

	sortModules(modules)

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

// sortModules sorts a given Module slice ascendingly by path.
// Main modules take precedence, so that they will represent the first elements of the sorted slice.
// If the path of two modules are equal, they'll be compared by their semantic version instead.
func sortModules(modules []Module) {
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

// Replacements may point to local directories, in which case their .Path is
// not the actual module's name, but the filepath as used in go.mod.
func resolveLocalModules(mainModuleDir string, modules []Module) error {
	for i, module := range modules {
		if module.Replace == nil {
			// Only replacements can be local
			continue
		}

		var localModuleDir string
		if filepath.IsAbs(module.Replace.Path) {
			localModuleDir = modules[i].Replace.Path
		} else { // Replacement path is relative to main module
			localModuleDir = filepath.Join(mainModuleDir, modules[i].Replace.Path)
		}

		if !IsModule(localModuleDir) {
			log.Warn().
				Str("dir", localModuleDir).
				Msg("local replacement is not a module")
			continue
		}

		err := resolveLocalModule(localModuleDir, modules[i].Replace)
		if err != nil {
			return fmt.Errorf("resolving local module %s failed: %w", module.Replace.Coordinates(), err)
		}
	}

	return nil
}

func resolveLocalModule(localModuleDir string, module *Module) error {
	if util.IsGoModule(module.Dir) && strings.HasPrefix(module.Dir, util.GetModuleCacheDir()) {
		// Module is in module cache
		return nil
	} else if !util.IsGoModule(localModuleDir) {
		return ErrNoModule
	}

	buf := new(bytes.Buffer)
	if err := gocmd.GetModule(localModuleDir, buf); err != nil {
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
			log.Warn().
				Err(err).
				Str("module", module.Path).
				Str("dir", localModuleDir).
				Msg("failed to resolve version of local module")
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
