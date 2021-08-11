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

package app

import (
	"errors"
	"flag"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

type Options struct {
	options.LogOptions
	options.OutputOptions
	options.SBOMOptions

	IncludeTest bool
	Main        string
	ModuleDir   string
}

func (o *Options) RegisterFlags(fs *flag.FlagSet) {
	o.LogOptions.RegisterFlags(fs)
	o.OutputOptions.RegisterFlags(fs)
	o.SBOMOptions.RegisterFlags(fs)

	fs.BoolVar(&o.IncludeTest, "test", false, "Include test dependencies")
	fs.StringVar(&o.Main, "main", "", "Path to the application's main package")
}

func (o Options) Validate() error {
	errs := make([]error, 0)

	if err := o.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := o.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	// TODO: verify that .Main exists and is relative to .ModuleDir

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}
