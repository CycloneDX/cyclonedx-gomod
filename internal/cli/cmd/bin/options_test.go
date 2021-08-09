package bin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBinOptions_Validate(t *testing.T) {
	t.Run("BinaryPath Not Exists", func(t *testing.T) {
		var binOptions BinOptions
		binOptions.BinaryPath = "./doesNotExist"

		err := binOptions.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "does not exist")
	})

	t.Run("BinaryPath Is Dir", func(t *testing.T) {
		var binOptions BinOptions
		binOptions.BinaryPath = "./"

		err := binOptions.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "is a directory")
	})
}
