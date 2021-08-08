package mod

import (
	"testing"

	"github.com/CycloneDX/cyclonedx-gomod/internal/cli/options"
	"github.com/stretchr/testify/require"
)

func TestOptions_Validate(t *testing.T) {
	t.Run("InvalidComponentType", func(t *testing.T) {
		var modOptions ModOptions
		modOptions.ComponentType = "foobar"

		err := modOptions.Validate()
		require.Error(t, err)

		var validationError *options.ValidationError
		require.ErrorAs(t, err, &validationError)

		require.Len(t, validationError.Errors, 1)
		require.Contains(t, validationError.Errors[0].Error(), "invalid component type")
	})
}
