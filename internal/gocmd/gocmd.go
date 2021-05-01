package gocmd

import (
	"fmt"
	"io"
	"os/exec"
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

// GetModules executes `go list -json -m all` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-list-m
func GetModules(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}

// GetVendoredModules executes `go mod vendor -v` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-vendor
func GetVendoredModules(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "vendor", "-v", "-e")
	cmd.Dir = modulePath
	cmd.Stderr = writer

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
