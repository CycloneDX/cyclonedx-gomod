package gomod

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
	"golang.org/x/mod/sumdb/dirhash"
)

var (
	// ErrNoGoModule indicates that a given path is not a valid Go module
	ErrNoGoModule = errors.New("not a Go module")

	// PrivateModulePatterns holds all module patterns provided via the
	// GONOPROXY and GOPRIVATE environment variables. If a module's path
	// matches any of these, cyclonedx-gomod won't reach out to external
	// data sources for it.
	//
	// See also:
	//  - https://golang.org/ref/mod#private-module-privacy
	//	- https://golang.org/ref/mod#environment-variables
	PrivateModulePatterns []string
)

func init() {
	patterns := make(map[string]bool)
	for _, pattern := range strings.Split(os.Getenv("GONOPROXY"), ",") {
		patterns[strings.TrimSpace(pattern)] = true
	}
	for _, pattern := range strings.Split(os.Getenv("GOPRIVATE"), ",") {
		patterns[strings.TrimSpace(pattern)] = true
	}

	PrivateModulePatterns = make([]string, 0, len(patterns))
	for pattern := range patterns {
		if pattern == "" {
			continue
		}
		PrivateModulePatterns = append(PrivateModulePatterns, pattern)
	}
}

// See https://golang.org/ref/mod#go-list-m
type Module struct {
	Dir     string
	Main    bool
	Path    string
	Replace *Module
	Version string

	Dependencies []*Module `json:"-"`
	Local        bool      `json:"-"`
	Vendored     bool      `json:"-"`
}

func (m Module) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}
	return m.Path + "@" + m.Version
}

func (m Module) Hash() (string, error) {
	h1, err := dirhash.HashDir(m.Dir, m.Coordinates(), dirhash.Hash1)
	if err != nil {
		return "", err
	}

	return h1, nil
}

func (m Module) Private() (private bool, err error) {
	for _, pattern := range PrivateModulePatterns {
		private, err = path.Match(pattern, m.Path)
		if private || err != nil {
			return
		}
	}
	return
}

func (m Module) PackageURL() string {
	return "pkg:golang/" + m.Coordinates()
}

func GetModules(path, pkgPath string) ([]Module, error) {
	if !util.IsGoModule(path) {
		return nil, ErrNoGoModule
	}

	var mainModule *Module
	var modules []Module
	var err error

	// We're going to call the go command a few times and
	// we'll (re-)use this buffer to write its output to.
	buf := new(bytes.Buffer)

	if !util.IsVendoring(path) {
		if err = gocmd.ListPackageDependencies(path, pkgPath, buf); err != nil {
			return nil, fmt.Errorf("listing modules failed: %w", err)
		}

		pkgs, err := parsePackages(buf)
		if err != nil {
			return nil, fmt.Errorf("parsing packages failed: %w", err)
		}

		uniqModules := make(map[string]Module)
		for i := range pkgs {
			if pkgs[i].Module == nil || pkgs[i].Standard {
				continue
			}

			if pkgs[i].Module.Main {
				if mainModule == nil {
					mainModule = pkgs[i].Module
				}
				continue
			}

			if _, ok := uniqModules[pkgs[i].Module.Coordinates()]; !ok {
				uniqModules[pkgs[i].Module.Coordinates()] = *pkgs[i].Module
			}
		}

		for coordinates := range uniqModules {
			modules = append(modules, uniqModules[coordinates])
		}
	} else { // TODO: Is this still necessary if `go list -deps` works for vendored modules as well?
		if err = gocmd.ListVendoredModules(path, buf); err != nil {
			return nil, fmt.Errorf("listing vendored modules failed: %w", err)
		}

		modules, err = parseVendoredModules(path, buf)
		if err != nil {
			return nil, fmt.Errorf("parsing vendored modules failed: %w", err)
		}

		// Main module is not included in vendored module list, so we have
		// to get it separately and prepend it to the module slice
		buf.Reset()
		if err = gocmd.GetModule(path, buf); err != nil {
			return nil, fmt.Errorf("listing main module failed: %w", err)
		}

		if err = json.NewDecoder(buf).Decode(mainModule); err != nil {
			return nil, fmt.Errorf("parsing main module failed: %w", err)
		}
	}

	// TODO: sort modules so their order is deterministic?

	if mainModule == nil {
		return nil, fmt.Errorf("failed to identify main module")
	}
	modules = append([]Module{*mainModule}, modules...)

	// Replacements may point to local directories, in which case their .Path is
	// not the actual module's name, but the filepath as used in go.mod.
	for i := range modules {
		if modules[i].Replace == nil {
			continue
		}

		localModulePath := ""
		if filepath.IsAbs(modules[i].Replace.Path) {
			localModulePath = modules[i].Replace.Path
		} else {
			localModulePath = filepath.Join(path, modules[i].Replace.Path)
		}
		if !util.IsGoModule(localModulePath) {
			continue
		}

		if err = resolveLocalModule(localModulePath, modules[i].Replace); err != nil {
			return nil, fmt.Errorf("resolving local module %s failed: %w", modules[i].Replace.Coordinates(), err)
		}
	}

	buf.Reset()
	if err = gocmd.GetModuleGraph(path, buf); err != nil {
		return nil, fmt.Errorf("listing module graph failed: %w", err)
	}

	if err = parseModuleGraph(buf, modules); err != nil {
		return nil, fmt.Errorf("parsing module graph failed: %w", err)
	}

	return modules, nil
}

// parseModules parses the output of `go list -json -m` into a Module slice
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

// parseVendoredModules parses the output of `go mod vendor -v` into a Module slice
func parseVendoredModules(path string, reader io.Reader) ([]Module, error) {
	modules := make([]Module, 0)

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "# ") {
			continue
		}

		fields := strings.Fields(strings.TrimPrefix(line, "# "))

		// Replacements may be specified as
		//   Path [Version] => Path [Version]
		arrowIndex := util.StringSliceIndex(fields, "=>")

		if arrowIndex == -1 {
			if len(fields) != 2 {
				return nil, fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
			}

			modules = append(modules, Module{
				Path:     fields[0],
				Version:  fields[1],
				Dir:      filepath.Join(path, "vendor", fields[0]),
				Vendored: true,
			})
		} else {
			pathParent := fields[0]
			versionParent := ""
			if arrowIndex == 2 {
				versionParent = fields[1]
			}

			pathReplacement := fields[arrowIndex+1]
			versionReplacement := ""
			if len(fields) == arrowIndex+3 {
				versionReplacement = fields[arrowIndex+2]
			}

			modules = append(modules, Module{
				Path:    pathParent,
				Version: versionParent,
				Replace: &Module{
					Path:     pathReplacement,
					Version:  versionReplacement,
					Dir:      filepath.Join(path, "vendor", pathParent), // Replacements are copied to their parents dir
					Vendored: true,
				},
			})
		}
	}

	return modules, nil
}

// parseModuleGraph parses the output of `go mod graph` and populates
// the .Dependencies field of a given Module slice.
//
// The Module slice is expected to contain only "effective" modules,
// with only a single version per module, as provided by `go list -m` or `go list -deps`.
func parseModuleGraph(reader io.Reader, modules []Module) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return fmt.Errorf("expected two fields per line, but got %d: %s", len(fields), line)
		}

		// The module graph contains dependency relationships for multiple versions of a module.
		// When identifying the ACTUAL dependant, we search for it in strict mode (versions must match).
		dependant := findModule(modules, fields[0], true)
		if dependant == nil {
			// TODO: log this in DEBUG level once we use a more sophisticated logger
			continue
		}

		// The identified module may depend on an older version of its dependency.
		// Due to Go's minimal version selection, that version may not be present in
		// the effective modules slice. Hence, we search for the dependency in non-strict mode.
		dependency := findModule(modules, fields[1], false)
		if dependency == nil {
			// TODO: log this in DEBUG level once we use a more sophisticated logger
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

func findModule(modules []Module, coordinates string, strict bool) *Module {
	for i := range modules {
		if coordinates == modules[i].Coordinates() || (!strict && strings.HasPrefix(coordinates, modules[i].Path+"@")) {
			if modules[i].Replace != nil {
				return modules[i].Replace
			}
			return &modules[i]
		}
	}
	return nil
}

func resolveLocalModule(localModulePath string, module *Module) error {
	if util.IsGoModule(module.Dir) && strings.HasPrefix(module.Dir, util.GetModuleCacheDir()) {
		// Module is in module cache
		return nil
	} else if !util.IsGoModule(localModulePath) {
		return ErrNoGoModule
	}

	buf := new(bytes.Buffer)
	if err := gocmd.GetModule(localModulePath, buf); err != nil {
		return err
	}
	localModule := new(Module)
	if err := json.NewDecoder(buf).Decode(localModule); err != nil {
		return err
	}

	module.Path = localModule.Path
	module.Local = true

	// Try to resolve the version. Only works when module.Dir is a Git repo.
	if module.Version == "" {
		version, err := GetModuleVersion(module.Dir)
		if err == nil {
			module.Version = version
		} else {
			// We don't fail with an error here, because our possibilities are limited.
			// module.Dir may be a Mercurial repo or just a normal directory, in which case we
			// cannot detect versions reliably right now.
			log.Printf("failed to resolve version of local module %s: %v\n", module.Path, err)
		}
	}

	return nil
}
