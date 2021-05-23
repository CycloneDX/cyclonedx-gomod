package gomod

import (
	"encoding/json"
	"errors"
	"io"
)

// Package represents parts of the struct that `go list` is working with.
// See https://golang.org/cmd/go/#hdr-List_packages_or_modules
type Package struct {
	ImportPath string  // directory containing package sources
	Name       string  // package name
	Standard   bool    // is this package part of the standard Go library?
	Module     *Module // info about package's containing module, if any (can be nil)
}

// parsePackages parses the output of `go list [-deps] -json` into a Package slice.
func parsePackages(reader io.Reader) ([]Package, error) {
	packages := make([]Package, 0)
	jsonDecoder := json.NewDecoder(reader)

	// Output is not a JSON array, so we have to parse one object after another
	for {
		var pkg Package
		if err := jsonDecoder.Decode(&pkg); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		packages = append(packages, pkg)
	}
	return packages, nil
}
