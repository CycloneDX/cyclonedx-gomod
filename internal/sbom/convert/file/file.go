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

package file

import (
	"fmt"

	"github.com/rs/zerolog/log"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
)

type Option func(absFilePath, relFilePath string, c *cdx.Component) error

func WithHashes(algos ...cdx.HashAlgorithm) Option {
	return func(abs, _ string, c *cdx.Component) error {
		hashes, err := sbom.CalculateFileHashes(abs, algos...)
		if err != nil {
			return err
		}

		c.Hashes = &hashes
		return nil
	}
}

func ToComponent(absFilePath, relFilePath string, options ...Option) (*cdx.Component, error) {
	log.Debug().
		Str("file", absFilePath).
		Msg("converting file to component")

	component := cdx.Component{
		Type:  cdx.ComponentTypeFile,
		Name:  relFilePath,
		Scope: cdx.ScopeRequired,
	}

	hashes, err := sbom.CalculateFileHashes(absFilePath, cdx.HashAlgoSHA1)
	if err != nil {
		return nil, err
	}
	component.Version = fmt.Sprintf("v0.0.0-%s", hashes[0].Value[:12])

	for _, option := range options {
		err = option(absFilePath, relFilePath, &component)
		if err != nil {
			return nil, err
		}
	}

	return &component, nil
}
