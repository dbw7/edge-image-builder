package kubernetes

import (
	"context"
	"fmt"
	config "github.com/suse-edge/edge-image-builder/pkg/config"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/http"
)

const (
	rke2InstallScriptURL = "https://get.rke2.io"
	k3sInstallScriptURL  = "https://get.k3s.io"
)

type ScriptDownloader struct{}

func (d ScriptDownloader) DownloadInstallScript(distribution, destinationPath string) (string, error) {
	var scriptURL string

	switch distribution {
	case config.KubernetesDistroRKE2:
		scriptURL = rke2InstallScriptURL
	case config.KubernetesDistroK3S:
		scriptURL = k3sInstallScriptURL
	default:
		return "", fmt.Errorf("unsupported distribution: %s", distribution)
	}

	installer := fmt.Sprintf("%s_installer.sh", distribution)
	destinationPath = filepath.Join(destinationPath, installer)

	if err := http.DownloadFile(context.Background(), scriptURL, destinationPath, nil); err != nil {
		return "", fmt.Errorf("downloading script: %w", err)
	}

	if err := os.Chmod(destinationPath, fileio.ExecutablePerms); err != nil {
		return "", fmt.Errorf("modifying script permissions: %w", err)
	}

	return installer, nil
}
