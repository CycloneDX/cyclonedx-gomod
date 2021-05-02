package main

import (
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/stretchr/testify/assert"
)

func TestValidateOptions(t *testing.T) {
	// Should fail on invalid ComponentType
	options := Options{
		ComponentTypeStr: "foobar",
	}
	err := validateOptions(&options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid component type")

	// Should set ComponentType when valid
	options = Options{
		ComponentTypeStr: "container",
	}
	err = validateOptions(&options)
	assert.NoError(t, err)
	assert.Equal(t, cdx.ComponentTypeContainer, options.ComponentType)

	// Should fail when invalid SerialNumber is provided
	options = Options{
		ComponentTypeStr: "container",
		SerialNumberStr:  "foobar",
	}
	err = validateOptions(&options)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid serial number")

	// Should not fail when invalid SerialNumber and NoSerialNumber are provided
	options = Options{
		ComponentTypeStr: "container",
		NoSerialNumber:   true,
		SerialNumberStr:  "foobar",
	}
	err = validateOptions(&options)
	assert.NoError(t, err)
	assert.Nil(t, options.SerialNumber)

	// Should set SerialNumber when provided an valid
	options = Options{
		ComponentTypeStr: "container",
		SerialNumberStr:  "b2330afe-e16b-4c4c-b10f-f571e96d6ecc",
	}
	err = validateOptions(&options)
	assert.NoError(t, err)
	assert.Equal(t, "b2330afe-e16b-4c4c-b10f-f571e96d6ecc", options.SerialNumber.String())
}
