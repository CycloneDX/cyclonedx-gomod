package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/version"
	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	itSnapshotter = cupaloy.NewDefaultConfig().
			WithOptions(cupaloy.SnapshotSubdirectory("./testdata/integration/snapshots"))

	// Prefix for temporary files and directories created during ITs
	tmpPrefix = version.Name + "_"

	// Serial number to use in order to keep generated SBOMs reproducible
	zeroUUID = uuid.MustParse("00000000-0000-0000-0000-000000000000")
)

// Integration test with a "simple" module with only a few dependencies,
// no replacements and no vendoring.
func TestIntegrationSimple(t *testing.T) {
	skipIfShort(t)

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom.xml")
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	err = executeCommand(Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/simple",
		OutputPath:      bomFile.Name(),
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

// Integration test with a module that uses replacement with a local module.
func TestIntegrationLocal(t *testing.T) {
	skipIfShort(t)

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom.xml")
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	err = executeCommand(Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/local",
		OutputPath:      bomFile.Name(),
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

// Integration test with a module that doesn't have any dependencies.
func TestIntegrationNoDependencies(t *testing.T) {
	skipIfShort(t)

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom.xml")
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	err = executeCommand(Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		IncludeStd:      true,
		ModulePath:      "./testdata/integration/no-dependencies",
		OutputPath:      bomFile.Name(),
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

// Integration test with a "simple" module with only a few dependencies,
// no replacements and no vendoring.
func TestIntegrationVendored(t *testing.T) {
	skipIfShort(t)

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom.xml")
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	err = executeCommand(Options{
		ComponentType:   cdx.ComponentTypeLibrary,
		ModulePath:      "./testdata/integration/vendored",
		OutputPath:      bomFile.Name(),
		ResolveLicenses: true,
		Reproducible:    true,
		SerialNumber:    &zeroUUID,
	})
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// Read SBOM and compare with snapshot
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func assertValidSBOM(t *testing.T, bomFilePath string) {
	inputFormat := "xml_v1_2"
	if strings.HasSuffix(bomFilePath, ".json") {
		inputFormat = "json_v1_2"
	}
	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFilePath, "--input-format", inputFormat, "--fail-on-errors")
	valOut, err := valCmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
	}
}
