package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"

	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
)

type Options struct {
	ComponentType    cdx.ComponentType
	ComponentTypeStr string
	IncludeStd       bool
	IncludeTest      bool
	ModulePath       string
	NoSerialNumber   bool
	NoVersionPrefix  bool
	OutputPath       string
	PackagePath      string
	ResolveLicenses  bool
	Reproducible     bool
	SerialNumber     *uuid.UUID
	SerialNumberStr  string
	ShowVersion      bool
	UseJSON          bool
}

func main() {
	var options Options

	flag.StringVar(&options.ComponentTypeStr, "type", string(cdx.ComponentTypeApplication), "Type of the main component")
	flag.BoolVar(&options.IncludeStd, "std", false, "Include Go standard library as component and dependency of the module")
	flag.BoolVar(&options.IncludeTest, "test", false, "Include test modules")
	flag.StringVar(&options.ModulePath, "module", ".", "Path to Go module")
	flag.BoolVar(&options.NoSerialNumber, "noserial", false, "Omit serial number")
	flag.BoolVar(&options.NoVersionPrefix, "novprefix", false, "Omit \"v\" version prefix")
	flag.StringVar(&options.OutputPath, "output", "-", "Output path")
	flag.StringVar(&options.PackagePath, "package", "...", "Package path") // TODO: add better usage text
	flag.BoolVar(&options.ResolveLicenses, "licenses", false, "Resolve module licenses")
	flag.BoolVar(&options.Reproducible, "reproducible", false, "Make the SBOM reproducible by omitting dynamic content")
	flag.StringVar(&options.SerialNumberStr, "serial", "", "Serial number (default [random UUID])")
	flag.BoolVar(&options.ShowVersion, "version", false, "Show version")
	flag.BoolVar(&options.UseJSON, "json", false, "Output in JSON format")
	flag.Parse()

	if options.ShowVersion {
		fmt.Println(version.Version)
		return
	}

	if err := validateOptions(&options); err != nil {
		log.Fatal(err)
	}

	if err := executeCommand(options); err != nil {
		log.Fatal(err)
	}
}

var allowedComponentTypes = []cdx.ComponentType{
	cdx.ComponentTypeApplication,
	cdx.ComponentTypeContainer,
	cdx.ComponentTypeDevice,
	cdx.ComponentTypeFile,
	cdx.ComponentTypeFirmware,
	cdx.ComponentTypeFramework,
	cdx.ComponentTypeLibrary,
	cdx.ComponentTypeOS,
}

func validateOptions(options *Options) error {
	isAllowedComponentType := false
	for i := range allowedComponentTypes {
		if allowedComponentTypes[i] == cdx.ComponentType(options.ComponentTypeStr) {
			isAllowedComponentType = true
			break
		}
	}
	if isAllowedComponentType {
		options.ComponentType = cdx.ComponentType(options.ComponentTypeStr)
	} else {
		return fmt.Errorf("invalid component type %s. See https://cyclonedx.org/docs/1.2/#type_classification", options.ComponentTypeStr)
	}

	// Serial numbers must be valid UUIDs
	if !options.NoSerialNumber && options.SerialNumberStr != "" {
		if serialNumber, err := uuid.Parse(options.SerialNumberStr); err != nil {
			return fmt.Errorf("invalid serial number: %w", err)
		} else {
			options.SerialNumber = &serialNumber
		}
	}

	if options.ModulePath == "." {
		abs, _ := filepath.Abs(options.ModulePath)
		options.ModulePath = abs
	}

	// TODO: verify that PackagePath is either "all", "...", relative to ModulePath or an absolute path to a subdirectory of ModulePath

	return nil
}

func executeCommand(options Options) error {
	log.Println("generating sbom")
	bom, err := sbom.Generate(options.ModulePath, sbom.GenerateOptions{
		ComponentType:   options.ComponentType,
		IncludeStdLib:   options.IncludeStd,
		IncludeTest:     options.IncludeTest,
		NoSerialNumber:  options.NoSerialNumber,
		NoVersionPrefix: options.NoVersionPrefix,
		PackagePath:     options.PackagePath,
		Reproducible:    options.Reproducible,
		ResolveLicenses: options.ResolveLicenses,
		SerialNumber:    options.SerialNumber,
	})
	if err != nil {
		return fmt.Errorf("generating sbom failed: %w", err)
	}

	var outputFormat cdx.BOMFileFormat
	if options.UseJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	log.Println("writing sbom")
	var outputWriter io.Writer
	if options.OutputPath == "" || options.OutputPath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(options.OutputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", options.OutputPath, err)
		}
		defer outputFile.Close()
		outputWriter = outputFile
	}

	encoder := cdx.NewBOMEncoder(outputWriter, outputFormat)
	encoder.SetPretty(true)

	if err = encoder.Encode(bom); err != nil {
		return fmt.Errorf("encoding BOM failed: %w", err)
	}

	return nil
}
