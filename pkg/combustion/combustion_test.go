package combustion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/suse-edge/edge-image-builder/pkg/config"
	"github.com/suse-edge/edge-image-builder/pkg/image"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupContext(t *testing.T) (ctx *config.Context, def *image.Definition, teardown func()) {
	configDir, err := os.MkdirTemp("", "eib-config-")
	require.NoError(t, err)

	buildDir, err := os.MkdirTemp("", "eib-build-")
	require.NoError(t, err)

	combustionDir, err := os.MkdirTemp("", "eib-combustion-")
	require.NoError(t, err)

	artefactsDir, err := os.MkdirTemp("", "eib-artefacts-")
	require.NoError(t, err)

	def = &image.Definition{}

	ctx = &config.Context{
		ImageConfigDir: configDir,
		BuildDir:       buildDir,
		CombustionDir:  combustionDir,
		ArtefactsDir:   artefactsDir,
		Definition:     def,
	}

	return ctx, def, func() {
		assert.NoError(t, os.RemoveAll(combustionDir))
		assert.NoError(t, os.RemoveAll(buildDir))
		assert.NoError(t, os.RemoveAll(artefactsDir))
		assert.NoError(t, os.RemoveAll(configDir))
	}
}

func TestGenerateComponentPath(t *testing.T) {
	// Setup
	ctx, _, teardown := setupContext(t)
	defer teardown()

	componentDir := filepath.Join(ctx.ImageConfigDir, "some-component")
	require.NoError(t, os.Mkdir(componentDir, 0o755))

	// Test
	generatedPath := generateComponentPath(ctx, "some-component")

	// Verify
	assert.Equal(t, componentDir, generatedPath)
}

func TestIsComponentConfigured(t *testing.T) {
	ctx, _, teardown := setupContext(t)
	defer teardown()

	componentDir := filepath.Join(ctx.ImageConfigDir, "existing-component")
	require.NoError(t, os.Mkdir(componentDir, 0o755))

	assert.True(t, isComponentConfigured(ctx, "existing-component"))
	assert.False(t, isComponentConfigured(ctx, "missing-component"))
	assert.False(t, isComponentConfigured(ctx, ""))
}
