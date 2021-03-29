package gomod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

var (
	ErrNoGoModule = errors.New("not a Go module")
)

// See https://golang.org/ref/mod#go-list-m
type Module struct {
	Dir     string
	Main    bool
	Path    string
	Replace *Module
	Version string

	Dependencies []*Module `json:"-"`
}

func (m Module) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}
	return m.Path + "@" + m.Version
}

func (m Module) PackageURL() string {
	return "pkg:golang/" + m.Coordinates()
}

func GetModules(path string) ([]Module, error) {
	if _, err := os.Stat(filepath.Join(path, "go.mod")); os.IsNotExist(err) {
		return nil, ErrNoGoModule
	}

	buffer := new(bytes.Buffer)
	if err := gocmd.GetModuleList(path, buffer); err != nil {
		return nil, err
	}

	modules, err := parseModules(buffer)
	if err != nil {
		return nil, err
	}

	buffer.Reset()
	if err = gocmd.GetModuleGraph(path, buffer); err != nil {
		return nil, err
	}

	if err = parseModuleGraph(buffer, modules); err != nil {
		return nil, err
	}

	return modules, nil
}

func parseModules(reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)
	jsonDecoder := json.NewDecoder(reader)

	// Output is not a JSON array, so we have to parse one object after another
	for {
		var mod Module
		if err := jsonDecoder.Decode(&mod); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		modules = append(modules, mod)
	}
	return modules, nil
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
			return fmt.Errorf("expected two fields per line, but got %d: %s", len(parts), line)
		}

		dependant := findModule(modules, parts[0])
		if dependant == nil {
			continue
		}

		dependency := findModule(modules, parts[1])
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

func findModule(modules []Module, coordinates string) *Module {
	for i := range modules {
		if coordinates == modules[i].Coordinates() {
			if modules[i].Replace != nil {
				return modules[i].Replace
			}
			return &modules[i]
		}
	}
	return nil
}
