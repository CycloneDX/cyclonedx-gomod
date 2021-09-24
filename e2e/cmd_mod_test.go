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

package e2e

import (
	"path/filepath"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	modcmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/mod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

// Integration test with a "simple" module with only a few dependencies,
// no replacements and no vendoring.
func TestModCmdSimple(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/simple.tar.gz")

	modOptions := modcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ComponentType: string(cdx.ComponentTypeLibrary),
		ModuleDir:     fixturePath,
	}

	runSnapshotIT(t, &modOptions.OutputOptions, func() error { return modcmd.Exec(modOptions) })
}

// Integration test with a module that uses replacement with a local module.
// The local dependency is not a Git repository and thus won't have a version.
func TestModCmdLocal(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/local.tar.gz")

	modOptions := modcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ComponentType: string(cdx.ComponentTypeLibrary),
		ModuleDir:     filepath.Join(fixturePath, "local"),
	}

	runSnapshotIT(t, &modOptions.OutputOptions, func() error { return modcmd.Exec(modOptions) })
}

// Integration test with a module that doesn't have any dependencies.
func TestModCmdNoDependencies(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/no-dependencies.tar.gz")

	modOptions := modcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ComponentType: string(cdx.ComponentTypeLibrary),
		ModuleDir:     fixturePath,
	}

	runSnapshotIT(t, &modOptions.OutputOptions, func() error { return modcmd.Exec(modOptions) })
}

// Integration test with a "simple" module with only a few dependencies,
// no replacements, but vendoring.
func TestModCmdVendored(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/vendored.tar.gz")

	modOptions := modcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ComponentType: string(cdx.ComponentTypeLibrary),
		ModuleDir:     fixturePath,
	}

	runSnapshotIT(t, &modOptions.OutputOptions, func() error { return modcmd.Exec(modOptions) })
}

// Integration test with a "simple" module with only a few dependencies,
// but as a subdirectory of a Git repository. The expectation is that the
// (pseudo-) version is inherited from the repository of the parent dir.
//
// nested/
// |-+ .git/
// |-+ simple/
//   |-+ go.mod
//   |-+ go.sum
//   |-+ main.go
func TestModCmdNested(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/nested.tar.gz")

	modOptions := modcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ComponentType: string(cdx.ComponentTypeLibrary),
		ModuleDir:     filepath.Join(fixturePath, "simple"),
	}

	runSnapshotIT(t, &modOptions.OutputOptions, func() error { return modcmd.Exec(modOptions) })
}
