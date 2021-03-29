package gomod

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPseudoVersion(t *testing.T) {
	version, err := GetPseudoVersion("../../")
	require.NoError(t, err)
	require.Regexp(t, "^v0\\.0\\.0-[\\d]{14}-[\\da-z]{12}$", version)
}
