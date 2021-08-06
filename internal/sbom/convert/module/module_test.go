package module

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveVCSURL(t *testing.T) {
	t.Run("GitHub", func(t *testing.T) {
		require.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", ResolveVCSURL("github.com/CycloneDX/cyclonedx-go"))
	})

	t.Run("GitHub with major version", func(t *testing.T) {
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", ResolveVCSURL("github.com/CycloneDX/cyclonedx-go/v2"))
		assert.Equal(t, "https://github.com/CycloneDX/cyclonedx-go", ResolveVCSURL("github.com/CycloneDX/cyclonedx-go/v222"))
	})

	t.Run("gopkg.in variant 1", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-playground/assert", ResolveVCSURL("gopkg.in/go-playground/assert.v1"))
	})

	t.Run("gopkg.in variant 2", func(t *testing.T) {
		require.Equal(t, "https://github.com/go-check/check", ResolveVCSURL("gopkg.in/check.v1"))
	})
}
