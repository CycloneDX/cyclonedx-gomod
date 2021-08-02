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

package convert

import (
	"regexp"
	"strings"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/license"
)

type Option func(gomod.Module, *cyclonedx.Component) error

// WithResolvedLicenses attempts to resolve licenses for the module
// and attach it to the component's license evidence.
// Unless failOnError is true, a resolution failure will no cause an error.
func WithResolvedLicenses(failOnError bool) Option {
	return func(m gomod.Module, c *cyclonedx.Component) error {
		resolvedLicenses, err := license.Resolve(m)
		if err == nil {
			componentLicenses := make(cyclonedx.Licenses, len(resolvedLicenses))
			for i := range resolvedLicenses {
				componentLicenses[i] = cyclonedx.LicenseChoice{License: &resolvedLicenses[i]}
			}
			c.Evidence = &cyclonedx.Evidence{
				Licenses: &componentLicenses,
			}
		} else if failOnError {
			return err
		}
		return nil
	}
}

// WithComponentType overrides the type of the component.
func WithComponentType(ctype cyclonedx.ComponentType) Option {
	return func(_ gomod.Module, c *cyclonedx.Component) error {
		c.Type = ctype
		return nil
	}
}

// WithTestScope overrides the scope of the component,
// if the corresponding module has the TestOnly flag set.
func WithTestScope(scope cyclonedx.Scope) Option {
	return func(m gomod.Module, c *cyclonedx.Component) error {
		if m.TestOnly {
			c.Scope = scope
		}
		return nil
	}
}

// ToComponent converts a gomod.Module to a CycloneDX component.
// The component can be further customized using options, before it's returned.
func ToComponent(module gomod.Module, options ...Option) (*cyclonedx.Component, error) {
	if module.Replace != nil {
		return ToComponent(*module.Replace, options...)
	}

	component := cyclonedx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cyclonedx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version,
		PackageURL: module.PackageURL(),
	}

	vcsURL := resolveVCSURL(module.Path)
	if vcsURL != "" {
		component.ExternalReferences = &[]cyclonedx.ExternalReference{
			{
				Type: cyclonedx.ERTypeVCS,
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

func ToComponents(modules []gomod.Module, options ...Option) ([]cyclonedx.Component, error) {
	components := make([]cyclonedx.Component, 0, len(modules))

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
