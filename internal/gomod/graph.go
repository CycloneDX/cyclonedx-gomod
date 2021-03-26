package gomod

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func (m Module) ModuleGraph() (map[string][]string, error) {
	cmd := exec.Command("go", "mod", "graph")
	cmd.Dir = m.Dir

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return m.parseModuleGraph(bytes.NewReader(output))
}

func (m Module) parseModuleGraph(reader io.Reader) (map[string][]string, error) {
	graph := make(map[string][]string)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 2 {
			return nil, fmt.Errorf("expected two fields per line, but got %d: %s", len(parts), line)
		}

		dependant := parts[0]
		if dependant == m.Path {
			// The main module has no version in the module graph
			dependant = m.Coordinates()
		}
		dependency := parts[1]

		dependencies, ok := graph[dependant]
		if !ok {
			dependencies = []string{dependency}
		} else {
			dependencies = append(dependencies, dependency)
		}
		graph[dependant] = dependencies

		// For a complete graph, dependencies must be included as dependants as well
		if _, ok := graph[dependency]; !ok {
			graph[dependency] = make([]string, 0)
		}
	}
	return graph, nil
}

func GetEffectiveModuleGraph(moduleGraph map[string][]string, modules []Module) (map[string][]string, error) {
	newGraph := make(map[string][]string)

	for dependant, dependencies := range moduleGraph {
		// Filter out dependants that haven't made it into the final module list
		moduleFound := false
		for _, module := range modules {
			if dependant == module.Coordinates() {
				// Handle replacement
				if module.Replace != nil {
					dependant = module.Replace.Coordinates()
				}
				moduleFound = true
				break
			}
		}
		if !moduleFound {
			continue
		}

		newGraph[dependant] = make([]string, len(dependencies))

		// Rewire dependencies so they point to the correct version
		for i := range dependencies {
			moduleFound := false
			for _, module := range modules {
				if strings.Index(dependencies[i], module.Path+"@") == 0 {
					// Handle replacement
					if module.Replace != nil {
						newGraph[dependant][i] = module.Replace.Coordinates()
					} else {
						newGraph[dependant][i] = module.Coordinates()
					}
					moduleFound = true
				}
			}
			if !moduleFound {
				return nil, fmt.Errorf("dependency %s does not exist in module list", dependencies[i])
			}
		}
	}

	return newGraph, nil
}
