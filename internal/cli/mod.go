package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/sbom"
	"github.com/google/uuid"
	"github.com/peterbourgon/ff/v3/ffcli"
)

type ModOptions struct {
	OutputOptions
	SBOMOptions

	IncludeTest     bool
	ModuleDir       string
	ResolveLicenses bool
}

func (m *ModOptions) RegisterFlags(fs *flag.FlagSet) {
	fs.BoolVar(&m.ResolveLicenses, "licenses", false, "Resolve module licenses")
	fs.BoolVar(&m.IncludeTest, "test", false, "Include test dependencies")
}

func newModCmd() *ffcli.Command {
	var modOptions ModOptions

	fs := flag.NewFlagSet("cyclonedx-gomod mod", flag.ExitOnError)
	modOptions.RegisterFlags(fs)
	modOptions.OutputOptions.RegisterFlags(fs)
	modOptions.SBOMOptions.RegisterFlags(fs)

	return &ffcli.Command{
		Name:       "mod",
		ShortHelp:  "Generate SBOM for a module",
		ShortUsage: "cyclonedx-gomod mod [FLAGS...] PATH",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 1 {
				return flag.ErrHelp
			}
			if len(args) == 0 {
				modOptions.ModuleDir = "."
			} else {
				modOptions.ModuleDir = args[0]
			}

			return execModCmd(modOptions)
		},
	}
}

func execModCmd(modOptions ModOptions) error {
	var err error
	var serial uuid.UUID
	if modOptions.SBOMOptions.SerialNumber != "" {
		if serial, err = uuid.Parse(modOptions.SBOMOptions.SerialNumber); err != nil {
			return err
		}
	}

	log.Println("generating sbom")
	bom, err := sbom.Generate(modOptions.ModuleDir, sbom.GenerateOptions{
		ComponentType:   cdx.ComponentTypeLibrary,
		IncludeStdLib:   false,
		IncludeTest:     modOptions.IncludeTest,
		NoSerialNumber:  modOptions.SBOMOptions.NoSerialNumber,
		NoVersionPrefix: modOptions.SBOMOptions.NoVersionPrefix,
		Reproducible:    modOptions.SBOMOptions.Reproducible,
		ResolveLicenses: modOptions.ResolveLicenses,
		SerialNumber:    &serial,
	})
	if err != nil {
		return fmt.Errorf("generating sbom failed: %w", err)
	}

	var outputFormat cdx.BOMFileFormat
	if modOptions.OutputOptions.UseJSON {
		outputFormat = cdx.BOMFileFormatJSON
	} else {
		outputFormat = cdx.BOMFileFormatXML
	}

	log.Println("writing sbom")
	var outputWriter io.Writer
	if modOptions.OutputOptions.FilePath == "" || modOptions.OutputOptions.FilePath == "-" {
		outputWriter = os.Stdout
	} else {
		outputFile, err := os.Create(modOptions.OutputOptions.FilePath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", modOptions.OutputOptions.FilePath, err)
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
