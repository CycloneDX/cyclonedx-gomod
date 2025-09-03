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

package bin

import (
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
)

// Option allows for customization of the generator using the
// functional options pattern.
type Option func(g *generator) error

// WithIncludeStdlib toggles the inclusion of a std component
// representing the Go standard library in the generated BOM.
//
// When enabled, the std component will be represented as a
// direct dependency of the main module.
func WithIncludeStdlib(enable bool) Option {
	return func(g *generator) error {
		g.includeStdlib = enable
		return nil
	}
}

// WithLicenseDetector sets the license detector.
//
// Because Go does not embed license information in binaries,
// performing license detection requires downloading of source code.
// This is done by `go mod download`ing all modules (including main)
// to the module cache.
//
// When nil, no license detection will be performed. Default is nil.
func WithLicenseDetector(detector licensedetect.Detector) Option {
	return func(g *generator) error {
		g.licenseDetector = detector
		return nil
	}
}

// WithLogger overrides the default logger of the generator.
func WithLogger(logger zerolog.Logger) Option {
	return func(g *generator) error {
		g.logger = logger
		return nil
	}
}

// WithVersionOverride overrides the version of the main component.
//
// This is useful in cases where a BOM is generated from development-
// or snapshot-builds, in which case Go will set the version of the main
// module to "(devel)". Because "(devel)" is very generic and not a valid semver,
// overriding it may be useful.
//
// For analyzing release builds, this option should be avoided.
func WithVersionOverride(version string) Option {
	return func(g *generator) error {
		g.versionOverride = version
		return nil
	}
}

// WithShortPURLS toggles the use of short PURLs without query parameters.
func WithShortPURLS(enable bool) Option {
	return func(g *generator) error {
		g.shortPURLs = enable
		return nil
	}
}
