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

import "github.com/rs/zerolog"

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

// WithLicenseDetection toggles the license detection feature.
func WithLicenseDetection(enable bool) Option {
	return func(g *generator) error {
		g.detectLicenses = enable
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
