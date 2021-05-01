package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"

	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
)

var (
	componentType   string
	includeStd      bool
	modulePath      string
	noSerialNumber  bool
	noVersionPrefix bool
	outputPath      string
	serialNumber    string
	showVersion     bool
	useJSON         bool

	allowedComponentTypes = []cdx.ComponentType{
		cdx.ComponentTypeApplication,
		cdx.ComponentTypeContainer,
		cdx.ComponentTypeDevice,
		cdx.ComponentTypeFile,
		cdx.ComponentTypeFirmware,
		cdx.ComponentTypeFramework,
		cdx.ComponentTypeLibrary,
		cdx.ComponentTypeOS,
	}
)

func main() {
	flag.StringVar(&componentType, "type", string(cdx.ComponentTypeApplication), "Type of the main component")
	flag.BoolVar(&includeStd, "std", false, "Include Go standard library as component and dependency of the module")
	flag.StringVar(&modulePath, "module", ".", "Path to Go module")
	flag.BoolVar(&noSerialNumber, "noserial", false, "Omit serial number")
	flag.BoolVar(&noVersionPrefix, "novprefix", false, "Omit \"v\" version prefix")
	flag.StringVar(&outputPath, "output", "-", "Output path")
	flag.StringVar(&serialNumber, "serial", "", "Serial number (default [random UUID])")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&useJSON, "json", false, "Output in JSON format")
	flag.Parse()

	if showVersion {
		fmt.Println(version.Version)
		return
	}

	if err := validateArguments(); err != nil {
		log.Fatal(err)
	}

	parsedSerialNumber := new(uuid.UUID)
	if serial, err := uuid.Parse(serialNumber); err == nil {
		parsedSerialNumber = &serial
	}

	bom, err := sbom.Generate(modulePath, sbom.GenerateOptions{
		ComponentType:   cdx.ComponentType(componentType),
		IncludeStdLib:   includeStd,
		NoSerialNumber:  noSerialNumber,
		NoVersionPrefix: noVersionPrefix,
		SerialNumber:    parsedSerialNumber,
	})

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

func validateArguments() error {
	isAllowedComponentType := false
	for i := range allowedComponentTypes {
		if allowedComponentTypes[i] == cdx.ComponentType(componentType) {
			isAllowedComponentType = true
			break
		}
	}
	if !isAllowedComponentType {
		return fmt.Errorf("invalid component type %s. See https://pkg.go.dev/github.com/CycloneDX/cyclonedx-go#ComponentType for options", componentType)
	}

	// Serial numbers must be valid UUIDs
	if !noSerialNumber && serialNumber != "" {
		if _, err := uuid.Parse(serialNumber); err != nil {
			return fmt.Errorf("invalid serial number: %w", err)
		}
	}

	return nil
}
