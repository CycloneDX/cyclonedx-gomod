package gocmd

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// GetVersion
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

// GetModuleList
func GetModuleList(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}

// GetModuleGraph
func GetModuleGraph(modulePath string, writer io.Writer) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = modulePath
	cmd.Stdout = writer

	return cmd.Run()
}
