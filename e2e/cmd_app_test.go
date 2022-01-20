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
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	appcmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/app"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

func TestAppCmdSimple(t *testing.T) {
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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

func TestAppCmdSimpleAssertLicenses(t *testing.T) {
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

	fixturePath := extractFixture(t, "./testdata/modcmd/simple.tar.gz")

	appOptions := appcmd.Options{
		SBOMOptions: options.SBOMOptions{
			AssertLicenses:  true,
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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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
	resetBuildEnv := setupAppTestBuildEnv(t)
	defer resetBuildEnv()

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

func setupAppTestBuildEnv(t *testing.T) func() {
	env, err := gocmd.GetEnv(zerolog.New(io.Discard))
	require.NoError(t, err)

	buildEnv := make(map[string][]string) // Env -> testValue, originalValue
	buildEnv["CGO_ENABLED"] = []string{"0", env["CGO_ENABLED"]}
	buildEnv["GOARCH"] = []string{"amd64", env["GOARCH"]}
	buildEnv["GOOS"] = []string{"linux", env["GOOS"]}
	buildEnv["GOFLAGS"] = []string{"-tags=foo,bar", env["GOFLAGS"]}
	// TODO: Figure out how to deal with GOVERSION which isn't overridable

	for k, vs := range buildEnv {
		os.Setenv(k, vs[0])
	}

	return func() {
		for k, vs := range buildEnv {
			os.Setenv(k, vs[1])
		}
	}
}
