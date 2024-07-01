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

package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
)

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// IsSubPath checks (lexically) if subject is a subpath of path.
func IsSubPath(subject, path string) (bool, error) {
	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return false, fmt.Errorf("failed to make %s absolute: %w", path, err)
	}

	subjectAbs, err := filepath.Abs(subject)
	if err != nil {
		return false, fmt.Errorf("failed to make %s absolute: %w", subject, err)
	}

	if !strings.HasPrefix(subjectAbs, pathAbs) {
		return false, nil
	}

	return true, nil
}

func ParseSpecVersion(specVersion string) (sv cdx.SpecVersion, err error) {
	switch specVersion {
	case cdx.SpecVersion1_0.String():
		sv = cdx.SpecVersion1_0
	case cdx.SpecVersion1_1.String():
		sv = cdx.SpecVersion1_1
	case cdx.SpecVersion1_2.String():
		sv = cdx.SpecVersion1_2
	case cdx.SpecVersion1_3.String():
		sv = cdx.SpecVersion1_3
	case cdx.SpecVersion1_4.String():
		sv = cdx.SpecVersion1_4
	case cdx.SpecVersion1_5.String():
		sv = cdx.SpecVersion1_5
	case cdx.SpecVersion1_6.String():
		sv = cdx.SpecVersion1_6
	default:
		err = cdx.ErrInvalidSpecVersion
	}

	return
}
