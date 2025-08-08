package image

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/suse-edge/edge-image-builder/pkg/context"
	"github.com/suse-edge/edge-image-builder/pkg/version"
	"gopkg.in/yaml.v3"
)

type Definition struct {
	APIVersion               string                           `yaml:"apiVersion"`
	Image                    context.Image                    `yaml:"image"`
	OperatingSystem          OperatingSystem                  `yaml:"operatingSystem"`
	EmbeddedArtifactRegistry context.EmbeddedArtifactRegistry `yaml:"embeddedArtifactRegistry"`
	Kubernetes               context.Kubernetes               `yaml:"kubernetes"`
}

type OperatingSystem struct {
	KernelArgs       []string                       `yaml:"kernelArgs"`
	Groups           []context.OperatingSystemGroup `yaml:"groups"`
	Users            []context.OperatingSystemUser  `yaml:"users"`
	Systemd          context.Systemd                `yaml:"systemd"`
	Suma             context.Suma                   `yaml:"suma"`
	Packages         context.Packages               `yaml:"packages"`
	IsoConfiguration context.IsoConfiguration       `yaml:"isoConfiguration"`
	RawConfiguration context.RawConfiguration       `yaml:"rawConfiguration"`
	Time             context.Time                   `yaml:"time"`
	Proxy            context.Proxy                  `yaml:"proxy"`
	Keymap           string                         `yaml:"keymap"`
	EnableFIPS       bool                           `yaml:"enableFIPS"`
}

var _ context.Definition = (*Definition)(nil)

func ParseImageDefinition(data []byte) (*Definition, error) {
	var definition Definition

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&definition); err != nil {
		return nil, fmt.Errorf("could not parse the image definition: %w", err)
	}
	definition.Image.ImageType = strings.ToLower(definition.Image.ImageType)

	if !version.IsSchemaVersionSupported(definition.APIVersion) {
		return nil, context.ErrorInvalidSchemaVersion
	}

	return &definition, nil
}

func (d *Definition) GetAPIVersion() string {
	return d.APIVersion
}

func (d *Definition) GetImage() context.Image {
	return d.Image
}

func (d *Definition) GetOperatingSystem() context.OperatingSystem {
	return &d.OperatingSystem
}

func (d *Definition) GetKubernetes() *context.Kubernetes {
	return &d.Kubernetes
}

func (d *Definition) GetEmbeddedArtifactRegistry() context.EmbeddedArtifactRegistry {
	return d.EmbeddedArtifactRegistry
}

func (o *OperatingSystem) GetUsers() []context.OperatingSystemUser {
	return o.Users
}

func (o *OperatingSystem) GetGroups() []context.OperatingSystemGroup {
	return o.Groups
}

func (o *OperatingSystem) GetSystemd() context.Systemd {
	return o.Systemd
}

func (o *OperatingSystem) GetSuma() context.Suma {
	return o.Suma
}

func (o *OperatingSystem) GetTime() context.Time {
	return o.Time
}

func (o *OperatingSystem) GetProxy() context.Proxy {
	return o.Proxy
}

func (o *OperatingSystem) GetKeymap() string {
	return o.Keymap
}

func (o *OperatingSystem) GetKernelArgs() []string {
	return o.KernelArgs
}

func (o *OperatingSystem) GetPackages() context.Packages {
	return o.Packages
}

func (o *OperatingSystem) GetEnableFIPS() bool {
	return o.EnableFIPS
}

func (o *OperatingSystem) GetIsoConfiguration() context.IsoConfiguration {
	return o.IsoConfiguration
}

func (o *OperatingSystem) GetRawConfiguration() context.RawConfiguration {
	return o.RawConfiguration
}
