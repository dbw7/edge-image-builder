package registry

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar/v3"
	context2 "github.com/suse-edge/edge-image-builder/pkg/context"
	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/http"
	"go.uber.org/zap"
)

type helmClient interface {
	AddRepo(repository *context2.HelmRepository) error
	RegistryLogin(repository *context2.HelmRepository) error
	Pull(chart string, repository *context2.HelmRepository, version, destDir string) (string, error)
	Template(chart, repository, version, valuesFilePath, kubeVersion, targetNamespace string, apiVersions []string) ([]map[string]any, error)
}

type helmChart struct {
	context2.HelmChart
	localPath     string
	repositoryURL string
}

type Registry struct {
	embeddedImages []context2.ContainerImage
	manifestsDir   string
	helmClient     helmClient
	helmCharts     []*helmChart
	helmValuesDir  string
	kubeVersion    string
}

func New(ctx *context2.Context, localManifestsDir string, helmClient helmClient, helmValuesDir string) (*Registry, error) {
	manifestsDir, err := storeManifests(ctx, localManifestsDir)
	if err != nil {
		return nil, fmt.Errorf("storing manifests: %w", err)
	}

	charts, err := storeHelmCharts(ctx, helmClient)
	if err != nil {
		return nil, fmt.Errorf("storing helm charts: %w", err)
	}

	return &Registry{
		embeddedImages: ctx.Definition.GetEmbeddedArtifactRegistry().ContainerImages,
		manifestsDir:   manifestsDir,
		helmClient:     helmClient,
		helmCharts:     charts,
		helmValuesDir:  helmValuesDir,
		kubeVersion:    ctx.Definition.GetKubernetes().Version,
	}, nil
}

func (r *Registry) ManifestsPath() string {
	return r.manifestsDir
}

func storeManifests(ctx *context2.Context, localManifestsDir string) (string, error) {
	const manifestsDir = "manifests"

	var manifestsPathPopulated bool

	manifestsDestDir := filepath.Join(ctx.BuildDir, manifestsDir)

	manifestURLs := ctx.Definition.GetKubernetes().Manifests.URLs
	if len(manifestURLs) != 0 {
		if err := os.MkdirAll(manifestsDestDir, os.ModePerm); err != nil {
			return "", fmt.Errorf("creating manifests dir: %w", err)
		}

		for index, manifestURL := range manifestURLs {
			filePath := filepath.Join(manifestsDestDir, fmt.Sprintf("dl-manifest-%d.yaml", index+1))

			if err := http.DownloadFile(context.Background(), manifestURL, filePath, nil); err != nil {
				return "", fmt.Errorf("downloading manifest '%s': %w", manifestURL, err)
			}
		}

		manifestsPathPopulated = true
	}

	if _, err := os.Stat(localManifestsDir); err == nil {
		if err = fileio.CopyFiles(localManifestsDir, manifestsDestDir, "", false, &fileio.NonExecutablePerms); err != nil {
			return "", fmt.Errorf("copying manifests: %w", err)
		}

		manifestsPathPopulated = true
	} else if !errors.Is(err, fs.ErrNotExist) {
		zap.S().Warnf("Searching for local manifests failed: %v", err)
	}

	if !manifestsPathPopulated {
		return "", nil
	}

	return manifestsDestDir, nil
}

func storeHelmCharts(ctx *context2.Context, helmClient helmClient) ([]*helmChart, error) {
	helm := ctx.Definition.GetKubernetes().Helm

	if len(helm.Charts) == 0 {
		return nil, nil
	}

	bar := progressbar.Default(int64(len(helm.Charts)), "Pulling selected Helm charts...")

	helmDir := filepath.Join(ctx.BuildDir, "helm")
	if err := os.MkdirAll(helmDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating helm directory: %w", err)
	}

	chartRepositories := mapChartsToRepos(&helm)

	var charts []*helmChart
	helmChartPaths := make(map[string]string)

	for i := range helm.Charts {
		chart := helm.Charts[i]
		chartID := fmt.Sprintf("%s-%s-%s", chart.RepositoryName, chart.Name, chart.Version)

		repository, ok := chartRepositories[helm.Charts[i].RepositoryName]
		if !ok {
			return nil, fmt.Errorf("repository not found for chart %s", helm.Charts[i].Name)
		}

		if _, exists := helmChartPaths[chartID]; !exists {
			localPath, err := downloadChart(helmClient, &helm.Charts[i], repository, helmDir)
			if err != nil {
				return nil, fmt.Errorf("downloading chart: %w", err)
			}

			helmChartPaths[chartID] = localPath
		}

		charts = append(charts, &helmChart{
			HelmChart:     helm.Charts[i],
			localPath:     helmChartPaths[chartID],
			repositoryURL: repository.URL,
		})

		_ = bar.Add(1)
	}

	return charts, nil
}

func mapChartsToRepos(helm *context2.Helm) map[string]*context2.HelmRepository {
	chartRepoMap := make(map[string]*context2.HelmRepository)

	for i := range helm.Charts {
		for j := range helm.Repositories {
			if helm.Charts[i].RepositoryName == helm.Repositories[j].Name {
				chartRepoMap[helm.Charts[i].RepositoryName] = &helm.Repositories[j]
			}
		}
	}

	return chartRepoMap
}

func downloadChart(helmClient helmClient, chart *context2.HelmChart, repo *context2.HelmRepository, destDir string) (string, error) {
	if strings.HasPrefix(repo.URL, "http") {
		if err := helmClient.AddRepo(repo); err != nil {
			return "", fmt.Errorf("adding repo: %w", err)
		}
	} else if repo.Authentication.Username != "" && repo.Authentication.Password != "" {
		if err := helmClient.RegistryLogin(repo); err != nil {
			return "", fmt.Errorf("logging into registry: %w", err)
		}
	}

	chartPath, err := helmClient.Pull(chart.Name, repo, chart.Version, destDir)
	if err != nil {
		return "", fmt.Errorf("pulling chart: %w", err)
	}

	return chartPath, nil
}

func (r *Registry) ContainerImages() ([]string, error) {
	manifestImages, err := r.manifestImages()
	if err != nil {
		return nil, fmt.Errorf("getting container images from manifests: %w", err)
	}

	chartImages, err := r.helmChartImages()
	if err != nil {
		return nil, fmt.Errorf("getting container images from helm charts: %w", err)
	}

	return deduplicateContainerImages(r.embeddedImages, manifestImages, chartImages), nil
}

func deduplicateContainerImages(embeddedImages []context2.ContainerImage, manifestImages, chartImages []string) []string {
	imageSet := map[string]bool{}

	for _, img := range embeddedImages {
		imageSet[img.Name] = true
	}

	for _, img := range manifestImages {
		imageSet[img] = true
	}

	for _, img := range chartImages {
		imageSet[img] = true
	}

	var images []string

	for img := range imageSet {
		images = append(images, img)
	}

	return images
}
