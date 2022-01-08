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

import "github.com/rs/zerolog"

// Option TODO
type Option func(g *generator) error

// WithLicenseResolution TODO
func WithLicenseResolution(enable bool) Option {
	return func(g *generator) error {
		g.resolveLicenses = enable
		return nil
	}
}

// WithLogger TODO
func WithLogger(logger zerolog.Logger) Option {
	return func(g *generator) error {
		g.logger = logger
		return nil
	}
}

// WithStdlib TODO
func WithStdlib(enable bool) Option {
	return func(g *generator) error {
		g.includeStdlib = enable
		return nil
	}
}

// WithVersionOverride TODO
func WithVersionOverride(version string) Option {
	return func(g *generator) error {
		g.versionOverride = version
		return nil
	}
}
