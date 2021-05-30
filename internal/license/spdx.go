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
// Copyright (c) Niklas Düster. All Rights Reserved.

package license

import (
	_ "embed"
	"encoding/json"
	"log"
)

var (
	//go:embed spdx-licenses.json
	licensesJSON []byte

	licenses []SPDXLicense
)

type SPDXLicense struct {
	ID        string `json:"licenseId"`
	Name      string `json:"name"`
	Reference string `json:"reference"`
}

func init() {
	licensesObj := struct {
		Licenses []SPDXLicense `json:"licenses"`
	}{}
	if err := json.Unmarshal(licensesJSON, &licensesObj); err != nil {
		log.Fatalf("failed to unmarshal SPDX license list: %v", err)
	}
	licenses = licensesObj.Licenses
}

func getLicenseByID(licenseID string) *SPDXLicense {
	for i := range licenses {
		if licenses[i].ID == licenseID {
			return &licenses[i]
		}
	}
	return nil
}
