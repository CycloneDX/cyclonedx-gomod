package util

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-go"
	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/stretchr/testify/require"
)

func TestSetSerialNumber(t *testing.T) {
	t.Run("NoSerialNumber", func(t *testing.T) {
		require.NoError(t, SetSerialNumber(nil, options.SBOMOptions{
			SerialNumber:   "",
			NoSerialNumber: true,
		}))
	})

	t.Run("DefaultRandomSerialNumber", func(t *testing.T) {
		bom := new(cyclonedx.BOM)

		require.NoError(t, SetSerialNumber(bom, options.SBOMOptions{
			SerialNumber:   "",
			NoSerialNumber: false,
		}))
		require.Regexp(t, `^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, bom.SerialNumber)
	})

	t.Run("CustomSerialNumber", func(t *testing.T) {
		bom := new(cyclonedx.BOM)

		require.NoError(t, SetSerialNumber(bom, options.SBOMOptions{
			SerialNumber:   "00000000-0000-0000-0000-000000000000",
			NoSerialNumber: false,
		}))
		require.Equal(t, "urn:uuid:00000000-0000-0000-0000-000000000000", bom.SerialNumber)
	})

	t.Run("InvalidCustomSerialNumber", func(t *testing.T) {
		require.Error(t, SetSerialNumber(new(cyclonedx.BOM), options.SBOMOptions{
			SerialNumber:   "invalid",
			NoSerialNumber: false,
		}))
	})
}
