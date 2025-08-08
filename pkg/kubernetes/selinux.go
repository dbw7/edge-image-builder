package kubernetes

import (
	"context"
	"fmt"
	config "github.com/suse-edge/edge-image-builder/pkg/config"
	"github.com/suse-edge/edge-image-builder/pkg/http"
	"path/filepath"
	"strings"
)

func SELinuxPackage(version string, sources *config.ArtifactSources) (string, error) {

	switch {
	case strings.Contains(version, config.KubernetesDistroK3S):
		return sources.Kubernetes.K3s.SELinuxPackage, nil
	case strings.Contains(version, config.KubernetesDistroRKE2):
		return sources.Kubernetes.Rke2.SELinuxPackage, nil
	default:
		return "", fmt.Errorf("invalid kubernetes version: %s", version)
	}
}

func SELinuxRepository(version string, sources *config.ArtifactSources) (config.AddRepo, error) {
	var url string

	switch {
	case strings.Contains(version, config.KubernetesDistroK3S):
		url = sources.Kubernetes.K3s.SELinuxRepository
	case strings.Contains(version, config.KubernetesDistroRKE2):
		url = sources.Kubernetes.Rke2.SELinuxRepository
	default:
		return config.AddRepo{}, fmt.Errorf("invalid kubernetes version: %s", version)
	}

	return config.AddRepo{
		URL:      url,
		Unsigned: true,
	}, nil
}

func DownloadSELinuxRPMsSigningKey(gpgKeysDir string) error {
	const rancherSigningKeyURL = "https://rpm.rancher.io/public.key"
	var signingKeyPath = filepath.Join(gpgKeysDir, "rancher-public.key")

	return http.DownloadFile(context.Background(), rancherSigningKeyURL, signingKeyPath, nil)
}
