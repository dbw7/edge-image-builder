package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	// Setup
	filename := "./testdata/valid_example.yaml"
	configData, err := os.ReadFile(filename)
	require.NoError(t, err)

	// Test
	imageConfig, err := Parse(configData)
	require.NoError(t, err)

	// Verify
	assert.Equal(t, "1.0", imageConfig.APIVersion)
	assert.Equal(t, "iso", imageConfig.Image.ImageType)
	assert.Equal(t, "slemicro5.5.iso", imageConfig.Image.BaseImage)
	assert.Equal(t, "eibimage.iso", imageConfig.Image.OutputImageName)
	assert.Equal(t, "https://elemental.docs.rancher.com", imageConfig.ElementalConfig.Elemental.Registration.RegistrationURL)
	assert.Equal(t, "sample-ca-cert", imageConfig.ElementalConfig.Elemental.Registration.CACert)
	assert.Equal(t, true, imageConfig.ElementalConfig.Elemental.Registration.EmulateTPM)
	assert.Equal(t, 1, imageConfig.ElementalConfig.Elemental.Registration.EmulateTPMSeed)
	assert.Equal(t, "tpm", imageConfig.ElementalConfig.Elemental.Registration.AuthType)

	expectedKernelArgs := []string{
		"alpha=foo",
		"beta=bar",
	}
	assert.Equal(t, expectedKernelArgs, imageConfig.OperatingSystem.KernelArgs)
}

func TestParseBadConfig(t *testing.T) {
	// Setup
	badData := []byte("Not actually YAML")

	// Test
	_, err := Parse(badData)

	// Verify
	require.Error(t, err)
	assert.ErrorContains(t, err, "could not parse the image configuration")
}
