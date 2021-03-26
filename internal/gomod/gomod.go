package gomod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
}

func (m Module) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}
	return m.Path + "@" + m.Version
}

func (m Module) PackageURL() string {
	return CoordinatesToPURL(m.Coordinates())
}

func GetModules(path string) ([]Module, error) {
	if _, err := os.Stat(filepath.Join(path, "go.mod")); os.IsNotExist(err) {
		return nil, ErrNoGoModule
	}

	cmd := exec.Command("go", "list", "-mod", "readonly", "-json", "-m", "all")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("executing command '%s' failed: %w", cmd.String(), err)
	}

	return parseModules(bytes.NewReader(output))
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

func CoordinatesToPURL(coordinates string) string {
	return "pkg:golang/" + coordinates
}
