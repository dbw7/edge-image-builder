package fileio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopyFile(t *testing.T) {
	const (
		source        = "file_io.go" // use the source code file as a valid input
		destDirPrefix = "eib-copy-file-test-"
	)

	tmpDir, err := os.MkdirTemp("", destDirPrefix)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		source      string
		destination string
		perms       os.FileMode
		expectedErr string
	}{
		{
			name:        "Source file does not exist",
			source:      "<missing>",
			expectedErr: "opening source file: open <missing>: no such file or directory",
		},
		{
			name:        "Destination is an empty file",
			source:      source,
			destination: "",
			expectedErr: "creating file with permissions: creating file: open : no such file or directory",
		},
		{
			name:        "Destination is a directory",
			source:      source,
			destination: tmpDir,
			expectedErr: fmt.Sprintf("creating file with permissions: creating file: open %s: is a directory", tmpDir),
		},
		{
			name:        "File is successfully copied",
			source:      source,
			destination: fmt.Sprintf("%s/copy.go", tmpDir),
			perms:       NonExecutablePerms,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CopyFile(test.source, test.destination, test.perms)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				require.Nil(t, err)

				src, err := os.ReadFile(test.source)
				require.NoError(t, err)

				dest, err := os.ReadFile(test.destination)
				require.NoError(t, err)
				assert.Equal(t, src, dest)

				info, err := os.Stat(test.destination)
				require.NoError(t, err)
				assert.Equal(t, test.perms, info.Mode())
			}
		})
	}
}

func TestCopyFileN(t *testing.T) {
	const (
		destDirPrefix  = "eib-copy-file-n-test-"
		srcFileContent = "CopyFileN test"
	)

	tmpDir, err := os.MkdirTemp("", destDirPrefix)
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	buffer := bytes.NewBufferString(srcFileContent)

	tests := []struct {
		name        string
		source      io.Reader
		destination string
		perms       os.FileMode
		expectedErr string
	}{
		{
			name:        "Destination is an empty file",
			source:      buffer,
			destination: "",
			expectedErr: "creating file with permissions: creating file: open : no such file or directory",
		},
		{
			name:        "Destination is a directory",
			source:      buffer,
			destination: tmpDir,
			expectedErr: fmt.Sprintf("creating file with permissions: creating file: open %s: is a directory", tmpDir),
		},
		{
			name:        "File is successfully copied",
			source:      buffer,
			destination: fmt.Sprintf("%s/copy", tmpDir),
			perms:       NonExecutablePerms,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CopyFileN(test.source, test.destination, test.perms, 1)

			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				require.Nil(t, err)

				dest, err := os.ReadFile(test.destination)
				require.NoError(t, err)
				assert.Equal(t, []byte(srcFileContent), dest)

				info, err := os.Stat(test.destination)
				require.NoError(t, err)
				assert.Equal(t, test.perms, info.Mode())
			}
		})
	}
}

func TestCopyFiles(t *testing.T) {
	const (
		expectedSubDirName = "sub1-copy-files"
	)

	pwd, err := os.Getwd()
	require.NoError(t, err)
	testDataPath := filepath.Join(pwd, "testdata", "copy-files")

	tests := []struct {
		name                     string
		expectedRootDirFileNames []string
		expectedSubDirFileNames  []string
		expectedPerms            []os.FileMode
		extension                string
		destDirPrefix            string
		copySubDir               bool
		perms                    *os.FileMode
	}{
		{
			name:                     "Copy full directory filesystem",
			expectedRootDirFileNames: []string{"gpg.gpg", "rpm.rpm", "sub1-copy-files", "unwritable.txt"},
			expectedSubDirFileNames:  []string{"dummy.txt", "gpg.gpg", "rpm.rpm"},
			expectedPerms:            []os.FileMode{NonExecutablePerms, 0o755},
			destDirPrefix:            "eib-copy-files-all-dirs-",
			copySubDir:               true,
			perms:                    &NonExecutablePerms,
		},
		{
			name:                     "Copy full directory structure and files with specific extension",
			expectedRootDirFileNames: []string{"rpm.rpm", "sub1-copy-files"},
			expectedSubDirFileNames:  []string{"rpm.rpm"},
			expectedPerms:            []os.FileMode{NonExecutablePerms, 0o755},
			destDirPrefix:            "eib-copy-files-ext-all-dirs-",
			extension:                ".rpm",
			copySubDir:               true,
			perms:                    &NonExecutablePerms,
		},
		{
			name:                     "Copy all files only from the root directory",
			expectedRootDirFileNames: []string{"gpg.gpg", "rpm.rpm", "unwritable.txt"},
			expectedPerms:            []os.FileMode{NonExecutablePerms},
			destDirPrefix:            "eib-copy-files-root-dir-only-",
			perms:                    &NonExecutablePerms,
		},
		{
			name:                     "Copy files with specific extension only from the root directory",
			expectedRootDirFileNames: []string{"rpm.rpm"},
			expectedPerms:            []os.FileMode{NonExecutablePerms},
			destDirPrefix:            "eib-copy-files-root-dir-only-",
			extension:                ".rpm",
			perms:                    &NonExecutablePerms,
		},
		{
			name:                     "Copy files while maintaining their original permissions only from root directory",
			expectedRootDirFileNames: []string{"gpg.gpg", "rpm.rpm", "unwritable.txt"},
			expectedPerms:            []os.FileMode{NonExecutablePerms, 0o444},
			destDirPrefix:            "eib-copy-files-keep-perms-root-dir-only-",
			perms:                    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rootDir, err := os.MkdirTemp("", test.destDirPrefix)
			require.NoError(t, err)

			err = CopyFiles(testDataPath, rootDir, test.extension, test.copySubDir, test.perms)
			require.NoError(t, err)

			if test.copySubDir {
				assertDir(t, rootDir, test.expectedRootDirFileNames, expectedSubDirName)
				assertDir(t, filepath.Join(rootDir, expectedSubDirName), test.expectedSubDirFileNames, "")
			} else {
				assertDir(t, rootDir, test.expectedRootDirFileNames, "")
			}

			assertPerms(t, rootDir, test.expectedPerms)

			err = os.RemoveAll(rootDir)
			require.NoError(t, err)
		})
	}
}

func TestCopyFilesMissingSource(t *testing.T) {
	err := CopyFiles("", "", "", false, &NonExecutablePerms)
	assert.EqualError(t, err, "reading source dir: open : no such file or directory")
}

func TestCopyFilesMissingDestination(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err)
	testDataPath := filepath.Join(pwd, "testdata", "copy-files")

	err = CopyFiles(testDataPath, "", "", false, &NonExecutablePerms)
	assert.EqualError(t, err, "creating directory '': mkdir : no such file or directory")
}

func assertDir(t *testing.T, dirPath string, expectedFileNames []string, expectedSubDirName string) {
	const (
		expectedFileContent = "copy-files-test-data"
	)

	rootDirFiles, err := os.ReadDir(dirPath)
	require.NoError(t, err)

	fileNames := []string{}
	for _, file := range rootDirFiles {
		fileNames = append(fileNames, file.Name())

		if expectedSubDirName == "" {
			assert.False(t, file.IsDir())
		}

		if file.IsDir() {
			assert.Equal(t, expectedSubDirName, file.Name())
			continue
		}

		fileContent, err := os.ReadFile(filepath.Join(dirPath, file.Name()))
		require.NoError(t, err)
		assert.Equal(t, []byte(expectedFileContent), fileContent)
	}

	assert.Equal(t, expectedFileNames, fileNames)
}

func assertPerms(t *testing.T, dirPath string, expectedPerms []os.FileMode) {
	rootDirFiles, err := os.ReadDir(dirPath)
	require.NoError(t, err)

	var filePerms []os.FileMode
	for _, file := range rootDirFiles {
		fileInfo, err := file.Info()
		require.NoError(t, err)

		filePerms = append(filePerms, fileInfo.Mode().Perm())
	}

	for _, perm := range filePerms {
		assert.Contains(t, expectedPerms, perm)
	}
}
