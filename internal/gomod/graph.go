package gomod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func getModuleGraph(path string, modules []Module) error {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return parseModuleGraph(bytes.NewReader(output), modules)
}

func parseModuleGraph(reader io.Reader, modules []Module) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			return fmt.Errorf("")
		}

		dependant := findModule(parts[0], modules)
		if dependant == nil {
			continue
		}

		dependency := findModule(parts[1], modules)
		if dependency == nil {
			continue
		}

		if dependant.Dependencies == nil {
			dependant.Dependencies = []*Module{dependency}
		} else {
			dependant.Dependencies = append(dependant.Dependencies, dependency)
		}
	}

	return nil
}
