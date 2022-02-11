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

package pkg

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
)

func TestToComponent(t *testing.T) {
	m := gomod.Module{
		Path:    "modulePath",
		Version: "moduleVersion",
	}

	p := gomod.Package{
		ImportPath: "packagePath",
	}

	c, err := ToComponent(zerolog.Nop(), p, m)
	require.NoError(t, err)
	require.Equal(t, "packagePath", c.Name)
	require.Equal(t, "moduleVersion", c.Version)
	require.Equal(t, "pkg:golang/packagePath@moduleVersion?type=package", c.PackageURL)
}
