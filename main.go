package main

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
)

var (
	modulePath     string
	noSerialNumber bool
	outputPath     string
	serialNumber   string
	showVersion    bool
	useJSON        bool
)

func main() {
	flag.StringVar(&modulePath, "module", ".", "Path to Go module")
	flag.BoolVar(&noSerialNumber, "noserial", false, "Omit serial number")
	flag.StringVar(&outputPath, "output", "-", "Output path")
	flag.StringVar(&serialNumber, "serial", "", "Serial number (default [random UUID])")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&useJSON, "json", false, "Output in JSON format")
	flag.Parse()

	if showVersion {
		fmt.Println(version.Version)
		return
	}

	modules, err := gomod.GetModules(modulePath)
	if err != nil {
		log.Fatalf("failed to get modules: %v", err)
	}

	for i := range modules {
		module := &modules[i]

		if module.Main && module.Version == "" {
			module.Version, _ = gomod.GetPseudoVersion(module.Dir)
		}

		module.Version = strings.Replace(module.Version, "+incompatible", "", -1)
	}

	mainModule := modules[0]
	modules = modules[1:]

	bom := cdx.NewBOM()
	if !noSerialNumber {
		if serialNumber == "" {
			bom.SerialNumber = uuid.New().URN()
		} else {
			bom.SerialNumber = "urn:uuid:" + serialNumber
		}
	}

	toolHashes, err := calculcateToolHashes()
	if err != nil {
		log.Fatalf("failed to calculate tool hashes: %v", err)
	}

	mainComponent := convertToComponent(mainModule)
	mainComponent.Scope = ""                          // Main component can't have a scope
	mainComponent.Type = cdx.ComponentTypeApplication // TODO: Make this configurable

	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().Format(time.RFC3339),
		Tools: &[]cdx.Tool{
			{
				Vendor:  version.Author,
				Name:    version.Name,
				Version: version.Version,
				Hashes:  &toolHashes,
			},
		},
		Component: &mainComponent,
	}

	components := make([]cdx.Component, len(modules))
	for i, module := range modules {
		components[i] = convertToComponent(module)
	}
	bom.Components = &components

	var outputFormat cdx.BOMFileFormat
	if useJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	var outputWriter io.Writer
	if outputPath == "" || outputPath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("failed to create output file %s: %v", outputPath, err)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	encoder := cdx.NewBOMEncoder(outputWriter, outputFormat)
	encoder.SetPretty(true)

	if err = encoder.Encode(bom); err != nil {
		log.Fatalf("encoding BOM failed: %v", err)
	}
}

func calculcateToolHashes() ([]cdx.Hash, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// TODO: exe might be a symlink

	exeFile, err := os.Open(exe)
	if err != nil {
		return nil, err
	}
	defer exeFile.Close()

	hashMD5 := md5.New()
	hashSHA1 := sha1.New()
	hashSHA256 := sha256.New()
	hashSHA512 := sha512.New()
	hashWriter := io.MultiWriter(hashMD5, hashSHA1, hashSHA256, hashSHA512)

	if _, err = io.Copy(hashWriter, exeFile); err != nil {
		return nil, err
	}

	return []cdx.Hash{
		{Algorithm: cdx.HashAlgoMD5, Value: fmt.Sprintf("%x", hashMD5.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA1, Value: fmt.Sprintf("%x", hashSHA1.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA256, Value: fmt.Sprintf("%x", hashSHA256.Sum(nil))},
		{Algorithm: cdx.HashAlgoSHA512, Value: fmt.Sprintf("%x", hashSHA512.Sum(nil))},
	}, nil
}

func convertToComponent(module gomod.Module) cdx.Component {
	if module.Replace != nil {
		replacementComponent := convertToComponent(*module.Replace)

		module.Replace = nil // Avoid endless recursion
		replacedComponent := convertToComponent(module)

		replacementComponent.Pedigree = &cdx.Pedigree{
			Ancestors: &[]cdx.Component{replacedComponent},
		}

		return replacementComponent
	}

	component := cdx.Component{
		BOMRef:     module.PackageURL(),
		Type:       cdx.ComponentTypeLibrary,
		Name:       module.Path,
		Version:    module.Version, // TODO: Make it configurable to strip the "v" prefix
		Scope:      cdx.ScopeRequired,
		PackageURL: module.PackageURL(),
	}

	// We currently don't have an accurate way of hashing the main module, as it may contain
	// files that are .gitignore'd and thus not part of the hashes in Go's sumdb.
	// Maybe we need to copy and modify the code from https://github.com/golang/mod/blob/release-branch.go1.15/sumdb/dirhash/hash.go
	if !module.Main {
		if hashes, err := module.Hashes(); err != nil {
			log.Fatalf("failed to calculate hashes for %s: %v", component.PackageURL, err)
		} else {
			component.Hashes = &hashes
		}
	}

	return component
}
