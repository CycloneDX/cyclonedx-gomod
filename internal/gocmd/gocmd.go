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

// GetModuleName returns the name of the module at a given path.
// See https://golang.org/ref/mod#go-list-m
func GetModuleName(modulePath string) (string, error) {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-m")
	cmd.Dir = modulePath

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// GetModuleList returns a list of all modules at a given path.
// See https://golang.org/ref/mod#go-list-m
func GetModuleList(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}

// GetModuleGraph returns the module graph for the module at a given path.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}
