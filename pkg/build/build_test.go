package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/image"
)

func TestSetupBuildDirectory_EmptyRootDir(t *testing.T) {
	buildDir, combustionDir, err := SetupBuildDirectory("")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(buildDir))
	}()

	require.DirExists(t, buildDir)
	require.DirExists(t, combustionDir)

	assert.Contains(t, buildDir, "build-")
	assert.Equal(t, filepath.Join(buildDir, "combustion"), combustionDir)
}

func TestSetupBuildDir_NonEmptyRootDir(t *testing.T) {
	tests := []struct {
		name    string
		rootDir string
	}{
		{
			name: "Existing root dir",
			rootDir: func() string {
				tmpDir, err := os.MkdirTemp("", "eib-test-")
				require.NoError(t, err)

				return tmpDir
			}(),
		},
		{
			name:    "Non-existing root dir",
			rootDir: "some-non-existing-dir",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				assert.NoError(t, os.RemoveAll(test.rootDir))
			}()

			buildDir, combustionDir, err := SetupBuildDirectory(test.rootDir)
			require.NoError(t, err)

			require.DirExists(t, buildDir)
			require.DirExists(t, combustionDir)

			assert.Contains(t, buildDir, filepath.Join(test.rootDir, "build-"))
			assert.Equal(t, filepath.Join(buildDir, "combustion"), combustionDir)
		})
	}
}

func TestGenerateBuildDirFilename(t *testing.T) {
	// Setup
	tmpDir, err := os.MkdirTemp("", "eib-")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	builder := Builder{
		context: &image.Context{
			BuildDir: tmpDir,
		},
	}

	testFilename := "build-dir-file.sh"

	// Test
	filename := builder.generateBuildDirFilename(testFilename)

	// Verify
	expectedFilename := filepath.Join(builder.context.BuildDir, testFilename)
	require.Equal(t, expectedFilename, filename)
}
