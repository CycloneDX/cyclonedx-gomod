package gocmd

import (
	"fmt"
	"io"
	"os"
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

// GetModuleGraph executes `go mod graph` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = modulePath
	cmd.Stdout = writer
	return cmd.Run()
}

// ModWhy executes `go mod why -m -vendor` and writes the output to a given writer.
// See https://golang.org/ref/mod#go-mod-why
func ModWhy(modulePath string, modules []string, writer io.Writer) error {
	args := []string{"mod", "why", "-m", "-vendor"}
	args = append(args, modules...)

	cmd := exec.Command("go", args...)
	cmd.Dir = modulePath
	cmd.Stderr = os.Stderr
	cmd.Stdout = writer
	return cmd.Run()
}
