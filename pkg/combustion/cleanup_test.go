package combustion

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/config"
)

func TestConfigureCleanupRaw(t *testing.T) {
	// Setup
	ctx, def, teardown := setupContext(t)
	defer teardown()

	def.Image.ImageType = config.TypeRAW
	ctx.Definition = def

	// Test
	scriptNames, err := configureCleanup(ctx)

	// Verify
	require.NoError(t, err)

	assert.Equal(t, []string{cleanupScriptName}, scriptNames)

	// -- Combustion Script
	expectedCombustionScript := filepath.Join(ctx.CombustionDir, cleanupScriptName)
	contents, err := os.ReadFile(expectedCombustionScript)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "rm -r /artefacts")
}

func TestConfigureCleanupISO(t *testing.T) {
	// Setup
	ctx, def, teardown := setupContext(t)
	defer teardown()

	def.Image.ImageType = config.TypeISO
	ctx.Definition = def

	// Test
	scriptNames, err := configureCleanup(ctx)

	// Verify
	require.NoError(t, err)

	assert.NotEqual(t, []string{cleanupScriptName}, scriptNames)

	// -- Combustion Script
	expectedCombustionScript := filepath.Join(ctx.CombustionDir, cleanupScriptName)
	assert.NoFileExists(t, expectedCombustionScript)
}
