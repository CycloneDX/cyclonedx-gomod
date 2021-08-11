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

package module

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/license"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom/convert/file"
	"github.com/rs/zerolog/log"
)

type Option func(gomod.Module, *cdx.Component) error

// WithLicenses attempts to resolve licenses for the module and attach them
// to the component's license evidence.
func WithLicenses() Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if m.Dir == "" {
			log.Warn().
				Str("module", m.Coordinates()).
				Str("reason", "module not in cache").
				Msg("can't resolve module license")
			return nil
		}

		resolvedLicenses, err := license.Resolve(m)

		if err == nil {
			componentLicenses := make(cdx.Licenses, len(resolvedLicenses))
			for i := range resolvedLicenses {
				componentLicenses[i] = cdx.LicenseChoice{License: &resolvedLicenses[i]}
			}

			c.Evidence = &cdx.Evidence{
				Licenses: &componentLicenses,
			}
		} else {
			if errors.Is(err, license.ErrNoLicenseDetected) {
				log.Warn().Str("module", m.Coordinates()).Msg("no license detected")
				return nil
			}

			return fmt.Errorf("failed to resolve license for %s: %v", m.Coordinates(), err)
		}

		return nil
	}
}

// WithLicensesMaybe is a conditional wrapper around WithLicenses.
// Passing false to this Option will effectively make it a no-op.
func WithLicensesMaybe(enabled bool) Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if enabled {
			return WithLicenses()(m, c)
		}

		return nil
	}
}

// WithComponentType overrides the type of the component.
func WithComponentType(ctype cdx.ComponentType) Option {
	return func(_ gomod.Module, c *cdx.Component) error {
		c.Type = ctype
		return nil
	}
}

func WithFiles() Option {
	return func(m gomod.Module, c *cdx.Component) error {
		var fileComponents []cdx.Component

		for _, filePath := range m.Files {
			fileComponent, err := file.ToComponent(filepath.Join(m.Dir, filePath), filePath,
				file.WithScope(cdx.ScopeRequired),
				file.WithHashes(cdx.HashAlgoMD5, cdx.HashAlgoSHA1, cdx.HashAlgoSHA256, cdx.HashAlgoSHA384, cdx.HashAlgoSHA512),
			)
			if err != nil {
				return err
			}

			fileComponents = append(fileComponents, *fileComponent)
		}

		if len(fileComponents) == 0 {
			return nil
		}

		c.Components = &fileComponents

		return nil
	}
}

// WithScope overrides the scope of the component.
func WithScope(scope cdx.Scope) Option {
	return func(m gomod.Module, c *cdx.Component) error {
		c.Scope = scope
		return nil
	}
}

// WithTestScope overrides the scope of the component,
// if the corresponding module has the TestOnly flag set.
func WithTestScope(scope cdx.Scope) Option {
	return func(m gomod.Module, c *cdx.Component) error {
		if m.TestOnly {
			c.Scope = scope
		}

		return nil
	}
}

// ToComponent converts a gomod.Module to a CycloneDX component.
// The component can be further customized using options, before it's returned.
func ToComponent(module gomod.Module, options ...Option) (*cdx.Component, error) {
	if module.Replace != nil {
		return ToComponent(*module.Replace, options...)
	}

	component := cdx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cdx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version,
		PackageURL: module.PackageURL(),
	}

	if module.TestOnly {
		component.Scope = cdx.ScopeOptional
	} else {
		component.Scope = cdx.ScopeRequired
	}

	vcsURL := resolveVCSURL(module.Path)
	if vcsURL != "" {
		component.ExternalReferences = &[]cdx.ExternalReference{
			{
				Type: cdx.ERTypeVCS,
				URL:  vcsURL,
			},
		}
	}

	for _, option := range options {
		if err := option(module, &component); err != nil {
			return nil, err
		}
	}

	return &component, nil
}

// ToComponents converts a slice of gomod.Module to a slice of CycloneDX components.
func ToComponents(modules []gomod.Module, options ...Option) ([]cdx.Component, error) {
	components := make([]cdx.Component, 0, len(modules))

	for i := range modules {
		component, err := ToComponent(modules[i], options...)
		if err != nil {
			return nil, err
		}
		components = append(components, *component)
	}

	return components, nil
}

var (
	// By convention, modules with a major version equal to or above v2
	// have it as suffix in their module path.
	vcsUrlMajorVersionSuffixRegex = regexp.MustCompile(`(/v[\d]+)$`)

	// gopkg.in with user segment
	// Example: gopkg.in/user/pkg.v3 -> github.com/user/pkg
	vcsUrlGoPkgInRegexWithUser = regexp.MustCompile(`^gopkg\.in/([^/]+)/([^.]+)\..*$`)

	// gopkg.in without user segment
	// Example: gopkg.in/pkg.v3 -> github.com/go-pkg/pkg
	vcsUrlGoPkgInRegexWithoutUser = regexp.MustCompile(`^gopkg\.in/([^.]+)\..*$`)
)

func resolveVCSURL(modulePath string) string {
	switch {
	case strings.HasPrefix(modulePath, "github.com/"):
		return "https://" + vcsUrlMajorVersionSuffixRegex.ReplaceAllString(modulePath, "")
	case vcsUrlGoPkgInRegexWithUser.MatchString(modulePath):
		return "https://" + vcsUrlGoPkgInRegexWithUser.ReplaceAllString(modulePath, "github.com/$1/$2")
	case vcsUrlGoPkgInRegexWithoutUser.MatchString(modulePath):
		return "https://" + vcsUrlGoPkgInRegexWithoutUser.ReplaceAllString(modulePath, "github.com/go-$1/$1")
	}

	return ""
}
