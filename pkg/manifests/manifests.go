package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/http"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

func GetAllImages(ctx *image.Context) ([]image.ContainerImage, error) {
	var downloadedManifestPaths []string
	var combinedManifestPaths []string
	var extractedImagesSet = make(map[string]string)
	var err error

	if len(ctx.ImageDefinition.Kubernetes.Manifests.URLs) != 0 {
		downloadDestination := filepath.Join(ctx.BuildDir, "downloaded-manifests")
		if err = os.MkdirAll(downloadDestination, os.ModePerm); err != nil {
			return nil, fmt.Errorf("creating %s dir: %w", downloadDestination, err)
		}

		downloadedManifestPaths, err = downloadManifests(ctx, downloadDestination)
		if err != nil {
			return nil, fmt.Errorf("error downloading manifests: %w", err)
		}
	}

	localManifestSrcDir := filepath.Join(ctx.ImageConfigDir, "kubernetes", "manifests")
	localManifestPaths, err := getLocalManifestPaths(localManifestSrcDir)
	if err != nil {
		return nil, fmt.Errorf("error getting local manifest paths: %w", err)
	}

	combinedManifestPaths = append(localManifestPaths, downloadedManifestPaths...)

	for _, manifestPath := range combinedManifestPaths {
		manifestData, err := readManifest(manifestPath)
		if err != nil {
			return nil, fmt.Errorf("error reading manifest %w", err)
		}

		storeManifestImageNames(manifestData, extractedImagesSet)
	}

	for _, definedImage := range ctx.ImageDefinition.EmbeddedArtifactRegistry.ContainerImages {
		extractedImagesSet[definedImage.Name] = definedImage.SupplyChainKey
	}

	allImages := make([]image.ContainerImage, 0, len(extractedImagesSet))
	for imageName, supplyChainKey := range extractedImagesSet {
		containerImage := image.ContainerImage{
			Name:           imageName,
			SupplyChainKey: supplyChainKey,
		}
		allImages = append(allImages, containerImage)
	}

	return allImages, nil
}

func readManifest(manifestPath string) (any, error) {
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("error reading manifest: %w", err)
	}

	if len(manifestData) == 0 {
		return nil, fmt.Errorf("invalid manifest")
	}

	var manifest any
	err = yaml.Unmarshal(manifestData, &manifest)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling manifest yaml '%s': %w", manifestPath, err)
	}

	return manifest, nil
}

func storeManifestImageNames(data any, imageSet map[string]string) {

	var findImages func(data any)
	findImages = func(data any) {
		switch t := data.(type) {
		case map[string]any:
			for k, v := range t {
				if k == "image" {
					if imageName, ok := v.(string); ok {
						imageSet[imageName] = ""
					}
				}
				findImages(v)
			}
		case []any:
			for _, v := range t {
				findImages(v)
			}
		}
	}

	findImages(data)
}

func getLocalManifestPaths(src string) ([]string, error) {
	if src == "" {
		return nil, fmt.Errorf("manifest source directory not defined")
	}

	var manifestPaths []string

	manifests, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("reading manifest source dir: %w", err)
	}

	for _, manifest := range manifests {
		manifestName := strings.ToLower(manifest.Name())
		if filepath.Ext(manifestName) != ".yaml" && filepath.Ext(manifestName) != ".yml" {
			zap.S().Warnf("Skipping %s as it is not a yaml file", manifest.Name())
			continue
		}

		sourcePath := filepath.Join(src, manifest.Name())
		manifestPaths = append(manifestPaths, sourcePath)
	}

	return manifestPaths, nil
}

func downloadManifests(ctx *image.Context, destPath string) ([]string, error) {
	manifests := ctx.ImageDefinition.Kubernetes.Manifests.URLs
	var manifestPaths []string

	for index, manifestURL := range manifests {
		filePath := filepath.Join(destPath, fmt.Sprintf("manifest-%d.yaml", index+1))
		manifestPaths = append(manifestPaths, filePath)

		if err := http.DownloadFile(context.Background(), manifestURL, filePath); err != nil {
			return nil, fmt.Errorf("downloading manifest '%s': %w", manifestURL, err)
		}
	}

	return manifestPaths, nil
}
