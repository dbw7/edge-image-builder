package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/config"
)

func TestWriteElemental(t *testing.T) {
	// Setup
	image := &config.ImageConfig{
		ElementalConfig: config.ElementalConfig{
			Elemental: config.Elemental{
				Registration: struct {
					RegistrationURL string `yaml:"url"`
					CACert          string `yaml:"ca-cert"`
					EmulateTPM      bool   `yaml:"emulate-tpm"`
					EmulateTPMSeed  int    `yaml:"emulated-tpm-seed"`
					AuthType        string `yaml:"auth"`
				}{
					RegistrationURL: "https://example.com/register",
					CACert:          "path/to/ca-cert.pem",
					EmulateTPM:      true,
					EmulateTPMSeed:  1,
					AuthType:        "tpm",
				},
			},
		},
	}

	context, err := NewContext("", "", true)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, CleanUpBuildDir(context))
	}()

	builder := Builder{
		context:     context,
		imageConfig: image,
	}

	builder.writeElementalConfig()

	expectedFilename := filepath.Join(context.CombustionDir, "elemental_config.yaml")
	foundBytes, err := os.ReadFile(expectedFilename)
	require.NoError(t, err)

	foundContents := string(foundBytes)
	assert.Contains(t, foundContents, "https://example.com/register")

}

func TestConfigureElementalScript(t *testing.T) {
	// Setup
	context, err := NewContext("", "", true)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, CleanUpBuildDir(context))
	}()

	builder := Builder{context: context}

	// Test
	err = builder.writeElementalScript()

	// Verify
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(builder.context.CombustionDir, modifyElementalScriptName))
	require.NoError(t, err)

	require.Equal(t, 1, len(builder.combustionScripts))
	assert.Equal(t, modifyElementalScriptName, builder.combustionScripts[0])
}
