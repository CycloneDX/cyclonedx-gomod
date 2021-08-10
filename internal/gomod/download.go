package gomod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gocmd"
)

// See https://golang.org/ref/mod#go-mod-download
type ModuleDownload struct {
	Path    string // module path
	Version string // module version
	Error   string // error loading module
	Dir     string // absolute path to cached source root directory
	Sum     string // checksum for path, version (as in go.sum)
}

func (m ModuleDownload) Coordinates() string {
	if m.Version == "" {
		return m.Path
	}

	return m.Path + "@" + m.Version
}

func Download(modules []Module) ([]ModuleDownload, error) {
	var downloads []ModuleDownload
	chunks := chunkModules(modules, 20)

	for _, chunk := range chunks {
		chunkDownloads, err := downloadInternal(chunk)
		if err != nil {
			return nil, err
		}

		downloads = append(downloads, chunkDownloads...)
	}

	return downloads, nil
}

func downloadInternal(modules []Module) ([]ModuleDownload, error) {
	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	coordinates := make([]string, len(modules))
	for i := range modules {
		coordinates[i] = modules[i].Coordinates()
	}

	err := gocmd.DownloadModules(coordinates, stdoutBuf, stderrBuf)
	if err != nil {
		// `go mod download` will exit with code 1 if *any* of the
		// module downloads failed. Download errors are reported for
		// each module separately via the .Error field (written to STDOUT).
		//
		// If a serious error occurred that prevented `go mod download`
		// from running alltogether, it's written to STDERR.
		//
		// See https://github.com/golang/go/issues/35380
		if stderrBuf.Len() != 0 {
			return nil, fmt.Errorf(stderrBuf.String())
		}
	}

	var downloads []ModuleDownload
	jsonDecoder := json.NewDecoder(stdoutBuf)

	for {
		var download ModuleDownload
		if err := jsonDecoder.Decode(&download); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		downloads = append(downloads, download)
	}

	return downloads, nil
}
