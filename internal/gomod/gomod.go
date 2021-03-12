package gomod

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"golang.org/x/mod/sumdb/dirhash"
)

var (
	ErrNoGitRepository = errors.New("not a git repository")
	ErrNoGoModule      = errors.New("not a Go module")
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
	return m.Path + "@" + m.Version
}

func (m Module) Hashes() ([]cdx.Hash, error) {
	if _, err := os.Stat(m.Dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("module dir %s does not exist", m.Dir)
	}

	h1, err := dirhash.HashDir(m.Dir, m.Path+"@"+m.Version, dirhash.Hash1)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate h1 hash: %w", err)
	}

	h1Bytes, err := base64.StdEncoding.DecodeString(h1[3:])
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode h1 hash: %w", err)
	}

	return []cdx.Hash{
		{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", h1Bytes)},
	}, nil
}

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
		// Skip empty lines
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

// GetPseudoVersion constructs a pseudo version for a Go module at the given path.
// Note that this is only possible when path points to a Git repository and the
// git binary is available in the system's PATH.
// See https://golang.org/ref/mod#pseudo-versions
func GetPseudoVersion(path string) (string, error) {
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return "", ErrNoGitRepository
	}

	cmd := exec.Command("git", "show", "-s", "--format=%H %cI")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command '%s' failed: %w", cmd.String(), err)
	}

	parts := strings.Fields(string(output))
	if len(parts) != 2 {
		return "", fmt.Errorf("expected two fields in git output, but got %d: %s", len(parts), output)
	}

	commitHash := parts[0][:12]
	commitDate, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to parse commit timestamp: %w", err)
	}

	return fmt.Sprintf("v0.0.0-%s-%s", commitDate.Format("20060102150405"), commitHash), nil
}

// GetVersionFromTag checks if the current commit is annotated with a tag and if yes, returns that tag's name.
// Note that this is only possible when path points to a Git repository and the
// git binary is available in the system's PATH.
func GetVersionFromTag(path string) (string, error) {
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return "", ErrNoGitRepository
	}

	cmd := exec.Command("git", "describe", "--exact-match", "--tags", "HEAD")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("executing command '%s' failed: %w", cmd.String(), err)
	}

	return strings.TrimSpace(string(output)), nil
}

func CoordinatesToPURL(coordinates string) string {
	return "pkg:golang/" + coordinates
}
