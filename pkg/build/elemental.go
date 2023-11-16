package build

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"gopkg.in/yaml.v3"
)

const (
	modifyElementalScriptName = "11-elemental.sh"
)

//go:embed scripts/elemental/11-elemental.sh
var modifyElementalScript string

func (b *Builder) writeElementalConfig() error {
	elementalFileDest := filepath.Join(b.context.CombustionDir, "elemental_config.yaml")
	yamlData, err := yaml.Marshal(&b.imageConfig.ElementalConfig)
	if err != nil {
		return fmt.Errorf("error writing elemental config: %w", err)

	}
	err = fileio.WriteFile(elementalFileDest, string(yamlData), nil)
	if err != nil {
		return fmt.Errorf("error writing elemental config: %w", err)

	}

	return nil
}

func (b *Builder) writeElementalScript() error {

	writtenFilename, err := b.writeCombustionFile(modifyElementalScriptName, modifyElementalScript, nil)
	if err != nil {
		return fmt.Errorf("writing elemental script: %w", err)
	}
	err = os.Chmod(writtenFilename, modifyScriptMode)
	if err != nil {
		return fmt.Errorf("adjusting permissions: %w", err)
	}

	b.registerCombustionScript(modifyElementalScriptName)

	return nil
}
