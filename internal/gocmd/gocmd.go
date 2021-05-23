package gocmd

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// GetVersion returns the version of Go in the environment.
func GetVersion() (string, error) {
	cmd := exec.Command("go", "version")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	fields := strings.Fields(string(output))
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
func GetModule(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m")
	cmd.Dir = modulePath
	cmd.Stdout = writer
	return cmd.Run()
}

// ListModules executes `go list -json -m all` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func ListModules(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = modulePath
	cmd.Stdout = writer
	return cmd.Run()
}

// ListVendoredModules executes `go mod vendor -v` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-vendor
func ListVendoredModules(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "vendor", "-v", "-e")
	cmd.Dir = modulePath
	cmd.Stderr = writer
	return cmd.Run()
}

// ListPackageDependencies executed `go list -deps -json` and writes the output to a given writer.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
func ListPackageDependencies(modulePath, packagePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-deps", "-json", filepath.Join(modulePath, packagePath))
	cmd.Dir = modulePath
	cmd.Stdout = writer
	return cmd.Run()
}

// GetModuleGraph executes `go mod graph` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = modulePath
	cmd.Stdout = writer
	return cmd.Run()
}
