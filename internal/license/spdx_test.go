package license

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLicenseByID(t *testing.T) {
	require.NotEmpty(t, licenses)

	license := getLicenseByID("MIT")
	require.NotNil(t, license)
	assert.Equal(t, "MIT", license.ID)
	assert.Equal(t, "MIT License", license.Name)
	assert.Equal(t, "https://spdx.org/licenses/MIT.html", license.Reference)

	license = getLicenseByID("doesNotExist")
	assert.Nil(t, license)
}
