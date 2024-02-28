package rancher

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/fileio"
	"github.com/suse-edge/edge-image-builder/pkg/http"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"github.com/suse-edge/edge-image-builder/pkg/registry"
	"github.com/suse-edge/edge-image-builder/pkg/template"
	"os"
	"path/filepath"
)

type Rancher struct {
	rancherDef    image.Rancher
	buildDir      string
	combustionDir string
	RancherImages []image.ContainerImage
}

//go:embed templates/helm-crd.yaml.tpl
var helmCRD string

func New(buildDir, combustionDir string, rancherDef image.Rancher, helm registry.Helm) (*Rancher, error) {
	r := Rancher{
		rancherDef:    rancherDef,
		buildDir:      buildDir,
		combustionDir: combustionDir,
	}

	if err := r.configureRancher(helm); err != nil {
		return nil, fmt.Errorf("configuring rancher: %w", err)
	}
	return &r, nil
}

func (r *Rancher) configureRancher(helm registry.Helm) error {
	rancherDir := filepath.Join(r.buildDir, "rancher")
	if err := os.Mkdir(rancherDir, os.ModePerm); err != nil {
		return fmt.Errorf("making rancher build dir: %w", err)
	}

	manifestDestDir := filepath.Join(r.combustionDir, "kubernetes", "manifests")
	if err := os.MkdirAll(manifestDestDir, os.ModePerm); err != nil {
		return fmt.Errorf("making kubernetes manifests destination dir: %w", err)
	}

	chartsDir := filepath.Join(r.buildDir, "component-charts")
	if err := os.MkdirAll(chartsDir, os.ModePerm); err != nil {
		return fmt.Errorf("creating component charts dir: %w", err)
	}

	if err := r.configureCertManager(manifestDestDir, chartsDir); err != nil {
		return fmt.Errorf("configuring cert manager: %w", err)
	}

	if err := helm.AddRepo("rancher", "https://releases.rancher.com/server-charts/stable"); err != nil {
		return fmt.Errorf("adding repo chart: %w", err)
	}

	chartPath, err := helm.Pull("rancher", "https://releases.rancher.com/server-charts/stable", r.rancherDef.Version, rancherDir)
	if err != nil {
		return fmt.Errorf("pulling chart: %w", err)
	}

	chartContent, err := registry.GetChartContent(chartPath)
	if err != nil {
		return fmt.Errorf("getting chart content: %w", err)
	}

	if err = r.writeRancherManifest(manifestDestDir, chartContent); err != nil {
		return fmt.Errorf("writing rancher manifest: %w", err)
	}

	rancherImagesPath := filepath.Join(rancherDir, "rancher-images.txt")

	if err = r.images(rancherImagesPath); err != nil {
		return fmt.Errorf("configuring rancher images: %w", err)
	}

	return nil
}

func (r *Rancher) configureCertManager(manifestDestDir, chartsDir string) error {
	certManagerPath := filepath.Join(manifestDestDir, "cert-manager-crds.yaml")
	certManagerURL := fmt.Sprintf("https://github.com/cert-manager/cert-manager/releases/download/%s/cert-manager.crds.yaml", r.rancherDef.CertManager.Version)
	if err := http.DownloadFile(context.Background(), certManagerURL, certManagerPath, nil); err != nil {
		return fmt.Errorf("downloading cert manager crds: %w", err)
	}

	if err := writeCertManagerManifest(chartsDir, r.rancherDef.CertManager); err != nil {
		return fmt.Errorf("writing cert manager manifest: %w", err)
	}

	return nil
}

func writeCertManagerManifest(chartsDir string, certManagerDef image.CertManager) error {
	certManagerFileName := "cert-manager-helm.yaml"
	certManagerFile := filepath.Join(chartsDir, certManagerFileName)
	certManager := struct {
		Name            string
		Namespace       string
		Repo            string
		Chart           string
		TargetNamespace string
		CreateNamespace bool
		Version         string
		ChartContent    string
		Set             map[string]any
	}{
		Name:            "cert-manager",
		Namespace:       "kube-system",
		Repo:            "https://charts.jetstack.io",
		Chart:           "cert-manager",
		TargetNamespace: "cert-manager",
		CreateNamespace: true,
		Version:         certManagerDef.Version,
	}
	data, err := template.Parse(certManagerFileName, helmCRD, certManager)
	if err != nil {
		return fmt.Errorf("applying template to %s: %w", certManagerFileName, err)
	}

	if err = os.WriteFile(certManagerFile, []byte(data), fileio.NonExecutablePerms); err != nil {
		return fmt.Errorf("writing file %s: %w", certManagerFileName, err)
	}

	return nil
}

func (r *Rancher) writeRancherManifest(manifestsDir, chartContent string) error {
	fmt.Println("chart content", chartContent)
	rancherFileName := "rancher-helm.yaml"
	rancherFile := filepath.Join(manifestsDir, rancherFileName)
	rancherDef := struct {
		Name            string
		Namespace       string
		Repo            string
		Chart           string
		TargetNamespace string
		CreateNamespace bool
		Version         string
		ChartContent    string
		Set             map[string]any
	}{
		Name:            "rancher",
		Namespace:       "kube-system",
		Repo:            "https://releases.rancher.com/server-charts/stable",
		Chart:           "rancher",
		TargetNamespace: "cattle-system",
		CreateNamespace: true,
		Version:         "v2.8.2",
		ChartContent:    chartContent,
		Set: map[string]any{
			"hostname": "https://192.168.1.213.sslip.io",
			//"rancherImage":          "127.0.0.1:6545/rancher/rancher",
			"systemDefaultRegistry": "127.0.0.1:6545",
			"useBundledSystemChart": true,
		},
	}
	data, err := template.Parse(rancherFileName, helmCRD, rancherDef)
	if err != nil {
		return fmt.Errorf("applying template to %s: %w", rancherFileName, err)
	}

	if err = os.WriteFile(rancherFile, []byte(data), fileio.NonExecutablePerms); err != nil {
		return fmt.Errorf("writing file %s: %w", rancherFileName, err)
	}

	return nil
}

func (r *Rancher) images(path string) error {
	rancherImagesURL := fmt.Sprintf("https://github.com/rancher/rancher/releases/download/%s/rancher-images.txt", r.rancherDef.Version)
	if err := http.DownloadFile(context.Background(), rancherImagesURL, path, nil); err != nil {
		return fmt.Errorf("downloading rancher images: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening rancher images: %w", err)
	}
	defer file.Close()

	var images []image.ContainerImage
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		images = append(images, image.ContainerImage{Name: scanner.Text()})
	}

	if err = scanner.Err(); err != nil {
		return fmt.Errorf("error reading rancher images file: %w", err)
	}

	r.RancherImages = images

	return nil
}
