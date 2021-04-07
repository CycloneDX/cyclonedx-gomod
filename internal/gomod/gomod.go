package gomod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/sumdb/dirhash"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
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

func (m Module) Hash() (string, error) {
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return "", fmt.Errorf("module dir %s does not exist", m.Dir)
	}

	h1, err := dirhash.HashDir(m.Dir, m.Coordinates(), dirhash.Hash1)
	if err != nil {
		return "", err
	}

	return h1, nil
}

func (m Module) PackageURL() string {
	return "pkg:golang/" + m.Coordinates()
}

func GetModules(path string) ([]Module, error) {
	if !util.IsGoModule(path) {
		return nil, ErrNoGoModule
	}

	buf := new(bytes.Buffer)
	if err := gocmd.GetModuleList(path, buf); err != nil {
		return nil, err
	}

	modules, err := parseModules(buf)
	if err != nil {
		return nil, err
	}

	// Replacements may point to local directories, in which case their .Path is
	// not the actual module's name, but the filepath as used in go.mod.
	for i := range modules {
		if modules[i].Replace == nil {
			continue
		}

		if err := resolveLocalModule(path, modules[i].Replace); err != nil {
			return nil, fmt.Errorf("resolving local module %s failed: %v", modules[i].Replace.Coordinates(), err)
		}
	}

	buf.Reset()
	if err = gocmd.GetModuleGraph(path, buf); err != nil {
		return nil, err
	}

	if err = parseModuleGraph(buf, modules); err != nil {
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

func resolveLocalModule(mainModulePath string, module *Module) error {
	modCacheDir := os.Getenv("GOMODCACHE")
	if modCacheDir == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = build.Default.GOPATH
		}
		modCacheDir = filepath.Join(gopath, "pkg", "mod")
	}

	if util.StartsWith(module.Dir, modCacheDir) {
		// Module is in module cache
		return nil
	}

	modulePath := ""
	if filepath.IsAbs(module.Path) {
		modulePath = module.Path
	} else {
		modulePath = filepath.Join(mainModulePath, module.Path)
	}
	if !util.IsGoModule(modulePath) {
		return ErrNoGoModule
	}

	buf := new(bytes.Buffer)
	if err := gocmd.GetModule(modulePath, buf); err != nil {
		return err
	}
	localModule := new(Module)
	if err := json.NewDecoder(buf).Decode(localModule); err != nil {
		return err
	}

	module.Path = localModule.Path
	// TODO: Resolve version. How can this be done when the local module isn't in a Git repo?
	return nil
}
