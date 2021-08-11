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

package gocmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// GetVersion returns the version of Go in the environment.
func GetVersion() (string, error) {
	buf := new(bytes.Buffer)
	if err := executeGoCommand([]string{"version"}, "", buf, nil); err != nil {
		return "", err
	}

	output := buf.String()
	fields := strings.Fields(output)
	if len(fields) != 4 {
		return "", fmt.Errorf("expected four fields in output, but got %d: %s", len(fields), output)
	}

	if fields[0] != "go" || fields[1] != "version" {
		return "", fmt.Errorf("unexpected output format: %s", output)
	}

	return fields[2], nil
}

// GetModule executes `go list -json -m` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func GetModule(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"list", "-mod", "readonly", "-json", "-m"}, moduleDir, writer, nil)
}

// ListModules executes `go list -json -m all` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func ListModules(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"list", "-mod", "readonly", "-json", "-m", "all"}, moduleDir, writer, nil)
}

// ListVendoredModules executes `go mod vendor -v` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-vendor
func ListVendoredModules(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"mod", "vendor", "-v", "-e"}, moduleDir, nil, writer)
}

// GetModuleGraph executes `go mod graph` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"mod", "graph"}, moduleDir, writer, nil)
}

// ModWhy executes `go mod why -m -vendor` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-why
func ModWhy(moduleDir string, modules []string, writer io.Writer) error {
	args := []string{"mod", "why", "-m", "-vendor"}
	args = append(args, modules...)
	return executeGoCommand(args, moduleDir, writer, os.Stderr)
}

// ModWhy executes `go version -m` and writes the output to a given writer.
func GetModulesFromBinary(binaryPath string, writer io.Writer) error {
	return executeGoCommand([]string{"version", "-m", binaryPath}, "", writer, nil)
}

// DownloadModules executes `go mod download -json` and writes the output to the given writers.
func DownloadModules(modules []string, stdout, stderr io.Writer) error {
	return executeGoCommand(append([]string{"mod", "download", "-json"}, modules...), os.TempDir(), stdout, stderr)
}

func executeGoCommand(args []string, dir string, stdout, stderr io.Writer) error {
	cmd := exec.Command("go", args...)
	log.Debug().Str("cmd", cmd.String()).Msg("executing command")

	if dir != "" {
		cmd.Dir = dir
	}

	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	return cmd.Run()
}
