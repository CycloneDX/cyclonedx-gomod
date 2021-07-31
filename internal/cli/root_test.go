package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecRootCmd(t *testing.T) {
	err := execRootCmd()
	require.Error(t, err)
	require.ErrorIs(t, err, flag.ErrHelp)
}
