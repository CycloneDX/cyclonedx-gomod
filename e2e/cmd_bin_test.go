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

	bincmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/bin"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

func TestBinCmdSimple(t *testing.T) {
	binOptions := bincmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		BinaryPath: "./testdata/bincmd/simple",
		Version:    "v1.0.0",
	}

	runSnapshotIT(t, &binOptions.OutputOptions, func() error { return bincmd.Exec(binOptions) })
}

func TestBinCmdSimple118(t *testing.T) {
	binOptions := bincmd.Options{
		SBOMOptions: options.SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		BinaryPath: "./testdata/bincmd/simple1.18",
	}

	runSnapshotIT(t, &binOptions.OutputOptions, func() error { return bincmd.Exec(binOptions) })
}

func TestBinCmdSimpleAssertLicenses(t *testing.T) {
	binOptions := bincmd.Options{
		SBOMOptions: options.SBOMOptions{
			AssertLicenses:  true,
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		BinaryPath: "./testdata/bincmd/simple",
		Version:    "v1.0.0",
	}

	runSnapshotIT(t, &binOptions.OutputOptions, func() error { return bincmd.Exec(binOptions) })
}
