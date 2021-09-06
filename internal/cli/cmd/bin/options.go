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

package bin

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

type Options struct {
	options.LogOptions
	options.OutputOptions
	options.SBOMOptions

	BinaryPath string
	Version    string
}

func (b *Options) RegisterFlags(fs *flag.FlagSet) {
	b.LogOptions.RegisterFlags(fs)
	b.OutputOptions.RegisterFlags(fs)
	b.SBOMOptions.RegisterFlags(fs)

	fs.StringVar(&b.Version, "version", "", "Version of the main component")
}

func (b Options) Validate() error {
	errs := make([]error, 0)

	if err := b.OutputOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}
	if err := b.SBOMOptions.Validate(); err != nil {
		var verr *options.ValidationError
		if errors.As(err, &verr) {
			errs = append(errs, verr.Errors...)
		} else {
			return err
		}
	}

	fileInfo, err := os.Stat(b.BinaryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("binary at %s does not exist", b.BinaryPath))
		} else {
			return err
		}
	}
	if fileInfo != nil && fileInfo.IsDir() {
		errs = append(errs, fmt.Errorf("%s is a directory", b.BinaryPath))
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}
