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
	"testing"

	appcmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/app"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

func TestAppCmdSimple(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/simple.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir: fixturePath,
		Main:      "",
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdSimpleWithFiles(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/simple.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		Main:            "",
		IncludeFiles:    true,
		IncludePackages: true,
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdSimpleWithPackages(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/simple.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		Main:            "",
		IncludePackages: true,
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdSimpleMultiCommandUUID(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/appcmd/simple-multi-command.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir: fixturePath,
		Main:      "cmd/uuid",
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdSimpleMultiCommandPURL(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/appcmd/simple-multi-command.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir: fixturePath,
		Main:      "cmd/purl",
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdVendored(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/vendored.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir: fixturePath,
		Main:      "",
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdVendoredWithFiles(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/vendored.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		Main:            "",
		IncludeFiles:    true,
		IncludePackages: true,
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}

func TestAppCmdVendoredWithPackages(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/modcmd/vendored.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		Main:            "",
		IncludePackages: true,
	}

	runSnapshotIT(t, &appOptions.OutputOptions, func() error { return appcmd.Exec(appOptions) })
}
