package e2e

import (
	"testing"

	bincmd "github.com/CycloneDX/cyclonedx-gomod/internal/cli/cmd/bin"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
)

func TestBinCmdSimple(t *testing.T) {
	binOptions := bincmd.BinOptions{
		SBOMOptions: options.SBOMOptions{
			Reproducible:    true,
			ResolveLicenses: true,
			SerialNumber:    zeroUUID.String(),
		},
		BinaryPath: "./testdata/bincmd/simple",
		Version:    "v1.0.0",
	}

	runSnapshotIT(t, &binOptions.OutputOptions, func() error { return bincmd.Exec(binOptions) })
}
