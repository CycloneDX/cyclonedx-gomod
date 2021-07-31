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

package model

import "fmt"

type Module struct {
	Path     string // module path
	Version  string // module version
	Checksum string // module checksum

	Main         bool      // is this the main module?
	Dependencies []*Module // modules this module depends on

	Local    bool // is this a local module?
	TestOnly bool // is this module only required for tests?
	Vendored bool // is this a vendored module?
}

func (m Module) Coordinates() string {
	if m.Version != "" {
		return fmt.Sprintf("%s@%s", m.Path, m.Version)
	}
	return m.Path
}

func (m Module) PackageURL() string {
	return fmt.Sprintf("pkg:golang/%s", m.Coordinates())
}
