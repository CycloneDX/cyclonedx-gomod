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

// GetModule writes the JSON representation of a given module to a writer.
// See https://golang.org/ref/mod#go-list-m
func GetModule(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}

// GetModuleList writes the JSON representation of all modules at a given path to a writer.
// See https://golang.org/ref/mod#go-list-m
func GetModuleList(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}

// GetModuleGraph writes the module graph for the module at a given path to a writer.
// See https://golang.org/ref/mod#go-mod-graph
func GetModuleGraph(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}
