package build

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
)

const (
	modifyRPMScriptName = "10-rpm-install.sh"
)

//go:embed scripts/rpms/10-rpm-install.sh.tpl
var modifyRPMScript string

func (b *Builder) processRPMs() error {
	rpmSourceDir, err := b.generateRPMPath()
	if err != nil {
		return fmt.Errorf("generating RPM path: %w", err)
	}
	// Only proceed with processing the RPMs if the directory exists
	if rpmSourceDir == "" {
		return nil
	}

	rpmFileNames, err := getRPMFileNames(rpmSourceDir)
	if err != nil {
		return fmt.Errorf("getting RPM file names: %w", err)
	}

	err = copyRPMs(rpmSourceDir, b.context.CombustionDir, rpmFileNames)
	if err != nil {
		return fmt.Errorf("copying RPMs over: %w", err)
	}

	err = b.writeRPMScript(rpmFileNames)
	if err != nil {
		return fmt.Errorf("writing the RPM install script %s: %w", modifyRPMScriptName, err)
	}

	return nil
}

func getRPMFileNames(rpmSourceDir string) ([]string, error) {
	var rpmFileNames []string

	rpms, err := os.ReadDir(rpmSourceDir)
	if err != nil {
		return nil, fmt.Errorf("reading RPM source dir: %w", err)
	}

	for _, rpmFile := range rpms {
		if filepath.Ext(rpmFile.Name()) == ".rpm" {
			rpmFileNames = append(rpmFileNames, rpmFile.Name())
		}
	}

	if len(rpmFileNames) == 0 {
		return nil, fmt.Errorf("no RPMs found")
	}

	return rpmFileNames, nil
}

func copyRPMs(rpmSourceDir string, rpmDestDir string, rpmFileNames []string) error {
	if rpmDestDir == "" {
		return fmt.Errorf("RPM destination directory cannot be empty")
	}
	for _, rpm := range rpmFileNames {
		sourcePath := filepath.Join(rpmSourceDir, rpm)
		destPath := filepath.Join(rpmDestDir, rpm)

		err := fileio.CopyFile(sourcePath, destPath)
		if err != nil {
			return fmt.Errorf("copying file %s: %w", sourcePath, err)
		}
	}

	return nil
}

func (b *Builder) writeRPMScript(rpmFileNames []string) error {
	values := struct {
		RPMs string
	}{
		RPMs: strings.Join(rpmFileNames, " "),
	}

	writtenFilename, err := b.writeCombustionFile(modifyRPMScriptName, modifyRPMScript, &values)
	if err != nil {
		return fmt.Errorf("writing RPM script: %w", err)
	}
	err = os.Chmod(writtenFilename, modifyScriptMode)
	if err != nil {
		return fmt.Errorf("adjusting permissions: %w", err)
	}

	b.registerCombustionScript(modifyRPMScriptName)

	return nil
}

func (b *Builder) generateRPMPath() (string, error) {
	rpmSourceDir := filepath.Join(b.context.ImageConfigDir, "rpms")
	_, err := os.Stat(rpmSourceDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("checking for RPM directory at %s: %w", rpmSourceDir, err)
	}

	return rpmSourceDir, nil
}
