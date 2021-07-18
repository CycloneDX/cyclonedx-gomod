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

package spdx

import (
	_ "embed"
)

//go:generate go run ../../../tools/spdx-gen/main.go -o spdx_gen.go
var licenses []License

type License struct {
	ID        string `json:"licenseId"`
	Reference string `json:"reference"`
}

func GetLicenseByID(licenseID string) *License {
	for i := range licenses {
		if licenses[i].ID == licenseID {
			return &licenses[i]
		}
	}
	return nil
}
