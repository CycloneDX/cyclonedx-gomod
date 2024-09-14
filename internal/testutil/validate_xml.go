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

package testutil

import (
	"fmt"
	cdx "github.com/CycloneDX/cyclonedx-go"
	"sync"

	"github.com/terminalstatic/go-xsd-validate"
)

var xmlSchemaFiles = map[cdx.SpecVersion]string{
	cdx.SpecVersion1_0: "./schema/bom-1.0.xsd",
	cdx.SpecVersion1_1: "./schema/bom-1.1.xsd",
	cdx.SpecVersion1_2: "./schema/bom-1.2.xsd",
	cdx.SpecVersion1_3: "./schema/bom-1.3.xsd",
	cdx.SpecVersion1_4: "./schema/bom-1.4.xsd",
	cdx.SpecVersion1_5: "./schema/bom-1.5.xsd",
	cdx.SpecVersion1_6: "./schema/bom-1.6.xsd",
}

var xsdValidateInitOnce sync.Once

type xmlValidator struct{}

func newXMLValidator() validator {
	var initErr error
	xsdValidateInitOnce.Do(func() {
		initErr = xsdvalidate.Init()
	})
	if initErr != nil {
		panic(initErr)
	}

	return &xmlValidator{}
}

func (xv xmlValidator) Validate(bom []byte, specVersion cdx.SpecVersion) error {
	schemaFilePath, ok := xmlSchemaFiles[specVersion]
	if !ok {
		return fmt.Errorf("no xml schema known for spec version %s", specVersion)
	}

	xsdHandler, err := xsdvalidate.NewXsdHandlerUrl(schemaFilePath, xsdvalidate.ParsErrVerbose)
	if err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}
	defer xsdHandler.Free()

	xmlHandler, err := xsdvalidate.NewXmlHandlerMem(bom, xsdvalidate.ParsErrVerbose)
	if err != nil {
		return fmt.Errorf("failed to parse bom xml: %w", err)
	}
	defer xmlHandler.Free()

	return xsdHandler.Validate(xmlHandler, xsdvalidate.ValidErrDefault)
}
