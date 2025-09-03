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

package app

import (
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
)

type Option func(g *generator) error

// WithIncludeFiles toggles the inclusion of files.
// Has no effect when packages are not included as well.
func WithIncludeFiles(enable bool) Option {
	return func(g *generator) error {
		g.includeFiles = enable
		return nil
	}
}

// WithIncludePackages toggles the inclusion of packages.
func WithIncludePackages(enable bool) Option {
	return func(g *generator) error {
		g.includePackages = enable
		return nil
	}
}

// WithIncludePaths toggles the inclusion of file paths.
// Has no effect when files are not included as well.
func WithIncludePaths(enable bool) Option {
	return func(g *generator) error {
		g.includePaths = enable
		return nil
	}
}

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

// WithMainDir overrides the main directory of the application.
func WithMainDir(dir string) Option {
	return func(g *generator) error {
		g.mainDir = dir
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
