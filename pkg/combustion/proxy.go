package combustion

import (
	_ "embed"
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"os"
	"path/filepath"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/log"
	"github.com/suse-edge/edge-image-builder/pkg/template"
)

const (
	proxyComponentName = "proxy"
	proxyScriptName    = "08-proxy-setup.sh"
)

//go:embed templates/08-proxy-setup.sh.tpl
var proxyScript string

func configureProxy(ctx *context.Context) ([]string, error) {
	proxy := ctx.Definition.GetOperatingSystem().GetProxy()
	if proxy.HTTPProxy == "" && proxy.HTTPSProxy == "" {
		log.AuditComponentSkipped(proxyComponentName)
		return nil, nil
	}

	if err := writeProxyCombustionScript(ctx); err != nil {
		log.AuditComponentFailed(proxyComponentName)
		return nil, err
	}

	log.AuditComponentSuccessful(proxyComponentName)
	return []string{proxyScriptName}, nil
}

func writeProxyCombustionScript(ctx *context.Context) error {
	proxyScriptFilename := filepath.Join(ctx.CombustionDir, proxyScriptName)

	values := struct {
		HTTPProxy  string
		HTTPSProxy string
		NoProxy    string
	}{
		HTTPProxy:  ctx.Definition.GetOperatingSystem().GetProxy().HTTPProxy,
		HTTPSProxy: ctx.Definition.GetOperatingSystem().GetProxy().HTTPSProxy,
		NoProxy:    strings.Join(ctx.Definition.GetOperatingSystem().GetProxy().NoProxy, ", "),
	}

	data, err := template.Parse(proxyScriptName, proxyScript, values)
	if err != nil {
		return fmt.Errorf("applying template to %s: %w", proxyScriptName, err)
	}

	if err := os.WriteFile(proxyScriptFilename, []byte(data), fileio.ExecutablePerms); err != nil {
		return fmt.Errorf("writing file %s: %w", proxyScriptFilename, err)
	}
	return nil
}
