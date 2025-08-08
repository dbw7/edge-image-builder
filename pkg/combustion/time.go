package combustion

import (
	_ "embed"
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/log"
	"github.com/suse-edge/edge-image-builder/pkg/template"
)

const (
	timeComponentName = "time"
	timeScriptName    = "11-time-setup.sh"
)

//go:embed templates/11-time-setup.sh.tpl
var timeScript string

func configureTime(ctx *context.Context) ([]string, error) {
	time := ctx.Definition.GetOperatingSystem().GetTime()
	if time.Timezone == "" {
		log.AuditComponentSkipped(timeComponentName)
		return nil, nil
	}

	if err := writeTimeCombustionScript(ctx); err != nil {
		log.AuditComponentFailed(timeComponentName)
		return nil, err
	}

	log.AuditComponentSuccessful(timeComponentName)
	return []string{timeScriptName}, nil
}

func writeTimeCombustionScript(ctx *context.Context) error {
	timeScriptFilename := filepath.Join(ctx.CombustionDir, timeScriptName)

	values := struct {
		Timezone  string
		Pools     []string
		Servers   []string
		ForceWait bool
	}{
		Timezone:  ctx.Definition.GetOperatingSystem().GetTime().Timezone,
		Pools:     ctx.Definition.GetOperatingSystem().GetTime().NtpConfiguration.Pools,
		Servers:   ctx.Definition.GetOperatingSystem().GetTime().NtpConfiguration.Servers,
		ForceWait: ctx.Definition.GetOperatingSystem().GetTime().NtpConfiguration.ForceWait,
	}

	data, err := template.Parse(timeScriptName, timeScript, values)
	if err != nil {
		return fmt.Errorf("applying template to %s: %w", timeScriptName, err)
	}

	if err := os.WriteFile(timeScriptFilename, []byte(data), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing file %s: %w", timeScriptFilename, err)
	}
	return nil
}
