package combustion

import (
	_ "embed"
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/log"
)

const (
	fipsComponentName = "fips"
	fipsScriptName    = "15-fips-setup.sh"
)

var (
	//go:embed templates/15-fips-setup.sh
	fipsScript     string
	FIPSPackages   = []string{"patterns-base-fips"}
	FIPSKernelArgs = []string{"fips=1"}
)

func configureFIPS(ctx *context.Context) ([]string, error) {
	fips := ctx.Definition.GetOperatingSystem().GetEnableFIPS()
	if !fips {
		log.AuditComponentSkipped(fipsComponentName)
		return nil, nil
	}

	if err := writeFIPSCombustionScript(ctx); err != nil {
		log.AuditComponentFailed(fipsComponentName)
		return nil, err
	}

	log.AuditComponentSuccessful(fipsComponentName)
	return []string{fipsScriptName}, nil
}

func writeFIPSCombustionScript(ctx *context.Context) error {
	fipsScriptFilename := filepath.Join(ctx.CombustionDir, fipsScriptName)

	if err := os.WriteFile(fipsScriptFilename, []byte(fipsScript), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing file %s: %w", fipsScriptFilename, err)
	}
	return nil
}
