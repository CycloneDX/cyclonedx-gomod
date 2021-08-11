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
	err := executeGoCommand([]string{"version"}, withStdout(buf))
	if err != nil {
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
	return executeGoCommand([]string{"list", "-mod", "readonly", "-json", "-m"}, withDir(moduleDir), withStdout(writer))
}

// ListModules executes `go list -json -m all` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func ListModules(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"list", "-mod", "readonly", "-json", "-m", "all"}, withDir(moduleDir), withStdout(writer))
}

// ListPackages executed `go list -deps -json` and writes the output to a given writer.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
func ListPackages(moduleDir, mainPackage string, writer io.Writer) error {
	if mainPackage == "" {
		mainPackage = "./..."
	}

	return executeGoCommand([]string{"list", "-deps", "-json", mainPackage},
		withDir(moduleDir),
		withStdout(writer),
		withStderr(os.Stderr),
	)
}

// ListVendoredModules executes `go mod vendor -v` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-vendor
func ListVendoredModules(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"mod", "vendor", "-v", "-e"}, withDir(moduleDir), withStderr(writer))
}

// GetModuleGraph executes `go mod graph` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(moduleDir string, writer io.Writer) error {
	return executeGoCommand([]string{"mod", "graph"}, withDir(moduleDir), withStdout(writer))
}

// ModWhy executes `go mod why -m -vendor` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-why
func ModWhy(moduleDir string, modules []string, writer io.Writer) error {
	return executeGoCommand(
		append([]string{"mod", "why", "-m", "-vendor"}, modules...),
		withDir(moduleDir),
		withStdout(writer),
		withStderr(os.Stderr), // reports download status
	)
}

// ModWhy executes `go version -m` and writes the output to a given writer.
func GetModulesFromBinary(binaryPath string, writer io.Writer) error {
	return executeGoCommand([]string{"version", "-m", binaryPath}, withStdout(writer))
}

// DownloadModules executes `go mod download -json` and writes the output to the given writers.
func DownloadModules(modules []string, stdout, stderr io.Writer) error {
	return executeGoCommand(
		append([]string{"mod", "download", "-json"}, modules...),
		withDir(os.TempDir()), // `mod download` modifies go.sum when executed in moduleDir
		withStdout(stdout),
		withStderr(stderr),
	)
}

type commandOption func(*exec.Cmd)

func withDir(dir string) commandOption {
	return func(c *exec.Cmd) {
		c.Dir = dir
	}
}

func withStderr(writer io.Writer) commandOption {
	return func(c *exec.Cmd) {
		c.Stderr = writer
	}
}

func withStdout(writer io.Writer) commandOption {
	return func(c *exec.Cmd) {
		c.Stdout = writer
	}
}

func executeGoCommand(args []string, options ...commandOption) error {
	cmd := exec.Command("go", args...)

	for _, option := range options {
		option(cmd)
	}

	log.Debug().
		Str("cmd", cmd.String()).
		Msg("executing command")

	return cmd.Run()
}
