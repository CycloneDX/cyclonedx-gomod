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

package main

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/stretchr/testify/assert"
)

func TestValidateOptions(t *testing.T) {
	t.Run("Invalid ComponentType", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "foobar",
		}
		err := validateOptions(&options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid component type")
	})

	t.Run("Valid ComponentType", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "container",
		}
		err := validateOptions(&options)
		assert.NoError(t, err)
		assert.Equal(t, cdx.ComponentTypeContainer, options.ComponentType)
	})

	t.Run("Invalid SerialNumber", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "container",
			SerialNumberStr:  "foobar",
		}
		err := validateOptions(&options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid serial number")
	})

	t.Run("Invalid SerialNumber with NoSerialNumber", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "container",
			NoSerialNumber:   true,
			SerialNumberStr:  "foobar",
		}
		err := validateOptions(&options)
		assert.NoError(t, err)
		assert.Nil(t, options.SerialNumber)
	})

	t.Run("Valid SerialNumber", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "container",
			SerialNumberStr:  "b2330afe-e16b-4c4c-b10f-f571e96d6ecc",
		}
		err := validateOptions(&options)
		assert.NoError(t, err)
		assert.Equal(t, "b2330afe-e16b-4c4c-b10f-f571e96d6ecc", options.SerialNumber.String())
	})

	t.Run("IncludeFiles without Distribution", func(t *testing.T) {
		options := Options{
			ComponentTypeStr: "container",
			Distribution:     false,
			NoSerialNumber:   true,
			IncludeFiles:     true,
		}
		err := validateOptions(&options)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be used without")
	})
}
