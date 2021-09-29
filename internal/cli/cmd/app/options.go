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
	"fmt"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"os"
	"path/filepath"
)

type Options struct {
	options.LogOptions
	options.OutputOptions
	options.SBOMOptions

	IncludeFiles bool
	Main         string
	ModuleDir    string
}

func (o *Options) RegisterFlags(fs *flag.FlagSet) {
	o.LogOptions.RegisterFlags(fs)
	o.OutputOptions.RegisterFlags(fs)
	o.SBOMOptions.RegisterFlags(fs)

	fs.BoolVar(&o.IncludeFiles, "files", false, "Include files")
	fs.StringVar(&o.Main, "main", "", "Path to the application's main package, relative to MODULE_PATH")
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

	err := o.validateMain(o.Main, &errs)
	if err != nil {
		return err
	}

	if len(errs) > 0 {
		return &options.ValidationError{Errors: errs}
	}

	return nil
}

func (o Options) validateMain(mainPkgDir string, errs *[]error) error {
	if filepath.IsAbs(mainPkgDir) {
		*errs = append(*errs, fmt.Errorf("main: must be a relative path"))
		return nil
	}

	mainPkgDir = filepath.Join(o.ModuleDir, mainPkgDir)
	isSubPath, err := util.IsSubPath(mainPkgDir, o.ModuleDir)
	if err != nil {
		return err
	}
	if !isSubPath {
		*errs = append(*errs, fmt.Errorf("main: must be a subpath of \"%s\"", o.ModuleDir))
		return nil
	}

	fileInfo, err := os.Stat(mainPkgDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			*errs = append(*errs, fmt.Errorf("main: \"%s\" does not exist", mainPkgDir))
			return nil
		}
		return err
	}
	if !fileInfo.IsDir() {
		*errs = append(*errs, fmt.Errorf("main: must be a directory, but \"%s\" is a file", mainPkgDir))
		return nil
	}

	pkg, err := gomod.LoadPackage(o.ModuleDir, o.Main)
	if err != nil {
		return fmt.Errorf("failed to load package: %w", err)
	}
	if pkg.Name != "main" {
		*errs = append(*errs, fmt.Errorf("main: must be main package, but is \"%s\"", pkg.Name))
		return nil
	}

	return nil
}
