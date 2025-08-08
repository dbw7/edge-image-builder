package kubernetes

import (
	"context"
	"fmt"
	context2 "github.com/suse-edge/edge-image-builder/pkg/context"
	"github.com/suse-edge/edge-image-builder/pkg/http"
	"path/filepath"
	"strings"
)

func SELinuxPackage(version string, sources *context2.ArtifactSources) (string, error) {

	switch {
	case strings.Contains(version, context2.KubernetesDistroK3S):
		return sources.Kubernetes.K3s.SELinuxPackage, nil
	case strings.Contains(version, context2.KubernetesDistroRKE2):
		return sources.Kubernetes.Rke2.SELinuxPackage, nil
	default:
		return "", fmt.Errorf("invalid kubernetes version: %s", version)
	}
}

func SELinuxRepository(version string, sources *context2.ArtifactSources) (context2.AddRepo, error) {
	var url string

	switch {
	case strings.Contains(version, context2.KubernetesDistroK3S):
		url = sources.Kubernetes.K3s.SELinuxRepository
	case strings.Contains(version, context2.KubernetesDistroRKE2):
		url = sources.Kubernetes.Rke2.SELinuxRepository
	default:
		return context2.AddRepo{}, fmt.Errorf("invalid kubernetes version: %s", version)
	}

	return context2.AddRepo{
		URL:      url,
		Unsigned: true,
	}, nil
}

func DownloadSELinuxRPMsSigningKey(gpgKeysDir string) error {
	const rancherSigningKeyURL = "https://rpm.rancher.io/public.key"
	var signingKeyPath = filepath.Join(gpgKeysDir, "rancher-public.key")

	return http.DownloadFile(context.Background(), rancherSigningKeyURL, signingKeyPath, nil)
}
