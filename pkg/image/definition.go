package image

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/config"
	"github.com/suse-edge/edge-image-builder/pkg/version"
	"gopkg.in/yaml.v3"
)

type Definition struct {
	APIVersion               string                          `yaml:"apiVersion"`
	Image                    config.Image                    `yaml:"image"`
	OperatingSystem          OperatingSystem                 `yaml:"operatingSystem"`
	EmbeddedArtifactRegistry config.EmbeddedArtifactRegistry `yaml:"embeddedArtifactRegistry"`
	Kubernetes               config.Kubernetes               `yaml:"kubernetes"`
}

type OperatingSystem struct {
	KernelArgs       []string                      `yaml:"kernelArgs"`
	Groups           []config.OperatingSystemGroup `yaml:"groups"`
	Users            []config.OperatingSystemUser  `yaml:"users"`
	Systemd          config.Systemd                `yaml:"systemd"`
	Suma             config.Suma                   `yaml:"suma"`
	Packages         config.Packages               `yaml:"packages"`
	IsoConfiguration config.IsoConfiguration       `yaml:"isoConfiguration"`
	RawConfiguration config.RawConfiguration       `yaml:"rawConfiguration"`
	Time             config.Time                   `yaml:"time"`
	Proxy            config.Proxy                  `yaml:"proxy"`
	Keymap           string                        `yaml:"keymap"`
	EnableFIPS       bool                          `yaml:"enableFIPS"`
}

var _ config.Definition = (*Definition)(nil)

func ParseImageDefinition(data []byte) (*Definition, error) {
	var definition Definition

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&definition); err != nil {
		return nil, fmt.Errorf("could not parse the image definition: %w", err)
	}
	definition.Image.ImageType = strings.ToLower(definition.Image.ImageType)

	if !version.IsSchemaVersionSupported(definition.APIVersion) {
		return nil, config.ErrorInvalidSchemaVersion
	}

	return &definition, nil
}

func (d *Definition) GetAPIVersion() string {
	return d.APIVersion
}

func (d *Definition) GetImage() config.Image {
	return d.Image
}

func (d *Definition) GetOperatingSystem() config.OperatingSystem {
	return &d.OperatingSystem
}

func (d *Definition) GetKubernetes() *config.Kubernetes {
	return &d.Kubernetes
}

func (d *Definition) GetEmbeddedArtifactRegistry() config.EmbeddedArtifactRegistry {
	return d.EmbeddedArtifactRegistry
}

func (o *OperatingSystem) GetUsers() []config.OperatingSystemUser {
	return o.Users
}

func (o *OperatingSystem) GetGroups() []config.OperatingSystemGroup {
	return o.Groups
}

func (o *OperatingSystem) GetSystemd() config.Systemd {
	return o.Systemd
}

func (o *OperatingSystem) GetSuma() config.Suma {
	return o.Suma
}

func (o *OperatingSystem) GetTime() config.Time {
	return o.Time
}

func (o *OperatingSystem) GetProxy() config.Proxy {
	return o.Proxy
}

func (o *OperatingSystem) GetKeymap() string {
	return o.Keymap
}

func (o *OperatingSystem) GetKernelArgs() []string {
	return o.KernelArgs
}

func (o *OperatingSystem) GetPackages() config.Packages {
	return o.Packages
}

func (o *OperatingSystem) GetEnableFIPS() bool {
	return o.EnableFIPS
}

func (o *OperatingSystem) GetIsoConfiguration() config.IsoConfiguration {
	return o.IsoConfiguration
}

func (o *OperatingSystem) GetRawConfiguration() config.RawConfiguration {
	return o.RawConfiguration
}
