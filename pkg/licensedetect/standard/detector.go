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

// Package standard provides the standard license detection logic.
//
// TODO: Find a better name? This one's too generic...
package standard

import (
	"errors"
	"fmt"
	"sort"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/go-enry/go-license-detector/v4/licensedb"
	"github.com/go-enry/go-license-detector/v4/licensedb/filer"
	"github.com/rs/zerolog"

	"github.com/CycloneDX/cyclonedx-gomod/pkg/licensedetect"
)

const minDetectionConfidence = 0.85

type detector struct {
	logger zerolog.Logger
}

// NewDetector TODO
func NewDetector(logger zerolog.Logger) licensedetect.Detector {
	return &detector{
		logger: logger,
	}
}

// Detect implements the licensedetect.Detector interface.
func (d detector) Detect(path, version, dir string) ([]cdx.License, error) {
	licensesFiler, err := filer.FromDirectory(dir)
	if err != nil {
		return nil, err
	}

	detectedLicenses, err := licensedb.Detect(licensesFiler)
	if err != nil {
		if errors.Is(err, licensedb.ErrNoLicenseFound) {
			return []cdx.License{}, nil
		}
		return nil, err
	}

	// Eventually we'll get multiple matches with equal confidence.
	// In order to at least make the output deterministic, we sort
	// the license names before iterating over the matches.
	// Maps are not sorted, so we need a temporary slice for this.
	detectedLicenseNames := make([]string, 0, len(detectedLicenses))
	for license := range detectedLicenses {
		detectedLicenseNames = append(detectedLicenseNames, license)
	}
	sort.Slice(detectedLicenseNames, func(i, j int) bool {
		return detectedLicenseNames[i] < detectedLicenseNames[j]
	})

	// Select the best match based on the highest confidence
	var detectedLicense string
	var detectedLicenseConfidence float32
	for _, license := range detectedLicenseNames {
		if detectedLicenses[license].Confidence > detectedLicenseConfidence {
			detectedLicense = license
			detectedLicenseConfidence = detectedLicenses[license].Confidence
		}
	}

	if detectedLicense == "" {
		return []cdx.License{}, nil
	}
	if detectedLicenseConfidence < minDetectionConfidence {
		d.logger.Debug().
			Str("module", fmt.Sprintf("%s@%s", path, version)).
			Str("license", detectedLicense).
			Float32("confidence", detectedLicenseConfidence).
			Float32("minConfidence", minDetectionConfidence).
			Msg("detection confidence for license is too low")
		return []cdx.License{}, nil
	}

	d.logger.Debug().
		Str("module", fmt.Sprintf("%s@%s", path, version)).
		Str("license", detectedLicense).
		Float32("confidence", detectedLicenseConfidence).
		Msg("license detected")

	return []cdx.License{
		{
			ID: detectedLicense,
		},
	}, nil
}
