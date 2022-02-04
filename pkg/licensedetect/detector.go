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

// Package licensedetect exposes cyclonedx-gomod's license detection functionality.
package licensedetect

import cdx "github.com/CycloneDX/cyclonedx-go"

// Detector is the interface that provides abstraction for license detection strategies.
//
// Detectors are provided with a module's path, version, and local directory (in Go's module cache).
// The latter may be empty, if the module has not been downloaded to the module cache.
type Detector interface {
	Detect(path, version, dir string) ([]cdx.License, error)
}
