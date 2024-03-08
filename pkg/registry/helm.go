package registry

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/suse-edge/edge-image-builder/pkg/image"
)

type HelmChart struct {
	CRD             HelmCRD
	ContainerImages []string
}

func HelmCharts(helmCharts []image.HelmChart, valuesDir, buildDir, kubeVersion string, helmClient image.HelmClient) ([]*HelmChart, error) {
	var charts []*HelmChart

	for _, helmChart := range helmCharts {
		c := helmChart
		chart, err := handleChart(&c, valuesDir, buildDir, kubeVersion, helmClient)
		if err != nil {
			return nil, fmt.Errorf("handling chart resource: %w", err)
		}

		charts = append(charts, chart)
	}

	return charts, nil
}

func handleChart(chart *image.HelmChart, valuesDir, buildDir, kubeVersion string, helmClient image.HelmClient) (*HelmChart, error) {
	var valuesPath string
	var valuesContent []byte
	if chart.ValuesFile != "" {
		var err error
		valuesPath = filepath.Join(valuesDir, chart.ValuesFile)
		valuesContent, err = os.ReadFile(valuesPath)
		if err != nil {
			return nil, fmt.Errorf("reading values content: %w", err)
		}
	}

	chartPath, err := downloadChart(chart, helmClient, buildDir)
	if err != nil {
		return nil, fmt.Errorf("downloading chart: %w", err)
	}

	images, err := getChartContainerImages(chart, helmClient, chartPath, valuesPath, kubeVersion)
	if err != nil {
		return nil, fmt.Errorf("getting chart container images: %w", err)
	}

	chartContent, err := getChartContent(chartPath)
	if err != nil {
		return nil, fmt.Errorf("getting chart content: %w", err)
	}

	helmChart := HelmChart{
		CRD:             NewHelmCRD(chart, chartContent, string(valuesContent)),
		ContainerImages: images,
	}

	return &helmChart, nil
}

func downloadChart(chart *image.HelmChart, helmClient image.HelmClient, destDir string) (string, error) {
	if err := helmClient.AddRepo(chart.Name, chart.Repo); err != nil {
		return "", fmt.Errorf("adding repo: %w", err)
	}

	chartPath, err := helmClient.Pull(chart.Name, chart.Repo, chart.Version, destDir)
	if err != nil {
		return "", fmt.Errorf("pulling chart: %w", err)
	}

	return chartPath, nil
}

func getChartContent(chartPath string) (string, error) {
	data, err := os.ReadFile(chartPath)
	if err != nil {
		return "", fmt.Errorf("reading chart: %w", err)
	}

	return base64.StdEncoding.EncodeToString(data), nil
}

func getChartContainerImages(chart *image.HelmChart, helmClient image.HelmClient, chartPath, valuesPath, kubeVersion string) ([]string, error) {
	chartResources, err := helmClient.Template(chart.Name, chartPath, chart.Version, valuesPath, kubeVersion)
	if err != nil {
		return nil, fmt.Errorf("templating chart: %w", err)
	}

	containerImages := map[string]bool{}
	for _, resource := range chartResources {
		storeManifestImages(resource, containerImages)
	}

	var images []string
	for i := range containerImages {
		images = append(images, i)
	}

	return images, nil
}
