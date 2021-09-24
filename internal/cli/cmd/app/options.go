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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
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
	fs.StringVar(&o.Main, "main", "main.go", "Path to the application's main file, relative to MODPATH")
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

func (o Options) validateMain(mainFilePath string, errs *[]error) error {
	mainFilePath = filepath.Join(o.ModuleDir, mainFilePath)

	if filepath.Ext(mainFilePath) != ".go" {
		*errs = append(*errs, fmt.Errorf("main: must be a go source file, but \"%s\" is not", mainFilePath))
		return nil
	}

	isSubPath, err := util.IsSubPath(mainFilePath, o.ModuleDir)
	if err != nil {
		return err
	}
	if !isSubPath {
		*errs = append(*errs, fmt.Errorf("main: must be a subpath of \"%s\"", o.ModuleDir))
		return nil
	}

	fileInfo, err := os.Stat(mainFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			*errs = append(*errs, fmt.Errorf("main: \"%s\" does not exist", mainFilePath))
			return nil
		}
		return err
	}

	if fileInfo.IsDir() {
		*errs = append(*errs, fmt.Errorf("main: must be a go file, but \"%s\" is a directory", mainFilePath))
		return nil
	}

	isMain, err := checkForMainPackage(mainFilePath)
	if err != nil {
		return err
	}
	if !isMain {
		*errs = append(*errs, fmt.Errorf("main: \"%s\" is not a main file", mainFilePath))
		return nil
	}

	return nil
}

func checkForMainPackage(filePath string) (bool, error) {
	mainFile, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer mainFile.Close()

	scanner := bufio.NewScanner(io.LimitReader(mainFile, 1024))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 && fields[0] == "package" && fields[1] == "main" {
			return true, nil
		}
	}

	return false, nil
}
