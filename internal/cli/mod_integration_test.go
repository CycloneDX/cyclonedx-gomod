package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/gomod"
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
	fixturePath := extractFixture(t, "./testdata/integration/simple.tar.gz")
	defer os.RemoveAll(fixturePath)

	runSnapshotIT(t, ModOptions{
		SBOMOptions: SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		ResolveLicenses: true,
	})
}

// Integration test with a module that uses replacement with a local module.
// The local dependency is not a Git repository and thus won't have a version.
func TestIntegrationLocal(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/integration/local.tar.gz")
	defer os.RemoveAll(fixturePath)

	runSnapshotIT(t, ModOptions{
		SBOMOptions: SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		ModuleDir:       filepath.Join(fixturePath, "local"),
		ResolveLicenses: true,
	})
}

// Integration test with a module that doesn't have any dependencies.
func TestIntegrationNoDependencies(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/integration/no-dependencies.tar.gz")
	defer os.RemoveAll(fixturePath)

	runSnapshotIT(t, ModOptions{
		SBOMOptions: SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		ResolveLicenses: true,
	})
}

// Integration test with a "simple" module with only a few dependencies,
// no replacements, but vendoring.
func TestIntegrationVendored(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/integration/vendored.tar.gz")
	defer os.RemoveAll(fixturePath)

	runSnapshotIT(t, ModOptions{
		SBOMOptions: SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		ModuleDir:       fixturePath,
		ResolveLicenses: true,
	})
}

// Integration test with a "simple" module with only a few dependencies,
// but as a subdirectory of a Git repository. The expectation is that the
// (pseudo-) version is inherited from the repository of the parent dir.
//
// nested/
// |-+ .git/
// |-+ simple/
//   |-+ go.mod
//   |-+ go.sum
//   |-+ main.go
func TestIntegrationNested(t *testing.T) {
	fixturePath := extractFixture(t, "./testdata/integration/nested.tar.gz")
	defer os.RemoveAll(fixturePath)

	runSnapshotIT(t, ModOptions{
		SBOMOptions: SBOMOptions{
			Reproducible: true,
			SerialNumber: zeroUUID.String(),
		},
		ModuleDir:       filepath.Join(fixturePath, "simple"),
		ResolveLicenses: true,
	})
}

func runSnapshotIT(t *testing.T, modOptions ModOptions) {
	skipIfShort(t)

	bomFileExtension := ".xml"
	if modOptions.OutputOptions.UseJSON {
		bomFileExtension = ".json"
	}

	// Create a temporary file to write the SBOM to
	bomFile, err := os.CreateTemp("", tmpPrefix+t.Name()+"_*.bom"+bomFileExtension)
	require.NoError(t, err)
	defer os.Remove(bomFile.Name())
	require.NoError(t, bomFile.Close())

	// Generate the SBOM
	modOptions.OutputOptions.FilePath = bomFile.Name()
	err = execModCmd(modOptions)
	require.NoError(t, err)

	// Sanity check: Make sure the SBOM is valid
	assertValidSBOM(t, bomFile.Name())

	// The versions of the modules in ./testdata are dynamic and depend on the current HEAD commit,
	// which would cause the snapshot comparisons to fail. Rest assured I felt dirty writing this.
	moduleVersion, err := gomod.GetModuleVersion(".")
	require.NoError(t, err)
	bomFileContent, err := os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	bomFileContent = regexp.MustCompile(`@?`+regexp.QuoteMeta(moduleVersion)).ReplaceAll(bomFileContent, nil)
	err = os.WriteFile(bomFile.Name(), bomFileContent, 0600)
	require.NoError(t, err)

	// Read SBOM and compare with snapshot
	bomFileContent, err = os.ReadFile(bomFile.Name())
	require.NoError(t, err)
	itSnapshotter.SnapshotT(t, string(bomFileContent))
}

func skipIfShort(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}

func extractFixture(t *testing.T, archivePath string) string {
	tmpDir, err := os.MkdirTemp("", tmpPrefix+t.Name()+"_*")
	require.NoError(t, err)

	cmd := exec.Command("tar", "xzf", archivePath, "-C", tmpDir)
	out, err := cmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(out))
	}

	return tmpDir
}

func assertValidSBOM(t *testing.T, bomFilePath string) {
	inputFormat := "xml_v1_3"
	if strings.HasSuffix(bomFilePath, ".json") {
		inputFormat = "json_v1_3"
	}
	valCmd := exec.Command("cyclonedx", "validate", "--input-file", bomFilePath, "--input-format", inputFormat, "--fail-on-errors")
	valOut, err := valCmd.CombinedOutput()
	if !assert.NoError(t, err) {
		// Provide some context when test is failing
		fmt.Printf("validation error: %s\n", string(valOut))
	}
}
