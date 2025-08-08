package image

import (
	"bytes"
	"fmt"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"strings"

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

func ParseImageDefinition(data []byte) (context.Definition, error) {
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

	return &ImageDefinitionAdapter{Definition: &definition}, nil
}

type ImageDefinitionAdapter struct {
	*Definition
}

func (a *ImageDefinitionAdapter) GetAPIVersion() string {
	return a.APIVersion
}

func (a *ImageDefinitionAdapter) GetImage() context.Image {
	return a.Image
}

func (a *ImageDefinitionAdapter) GetOperatingSystem() context.OperatingSystemInterface {
	return &ImageOSAdapter{OS: &a.OperatingSystem}
}

func (a *ImageDefinitionAdapter) GetKubernetes() *context.Kubernetes {
	return &a.Kubernetes
}

func (a *ImageDefinitionAdapter) GetEmbeddedArtifactRegistry() context.EmbeddedArtifactRegistry {
	return a.EmbeddedArtifactRegistry
}

type ImageOSAdapter struct {
	OS *OperatingSystem
}

func (o *ImageOSAdapter) GetUsers() []context.OperatingSystemUser {
	return o.OS.Users
}

func (o *ImageOSAdapter) GetGroups() []context.OperatingSystemGroup {
	return o.OS.Groups
}

func (o *ImageOSAdapter) GetSystemd() context.Systemd {
	return o.OS.Systemd
}

func (o *ImageOSAdapter) GetSuma() context.Suma {
	return o.OS.Suma
}

func (o *ImageOSAdapter) GetTime() context.Time {
	return o.OS.Time
}

func (o *ImageOSAdapter) GetProxy() context.Proxy {
	return o.OS.Proxy
}

func (o *ImageOSAdapter) GetKeymap() string {
	return o.OS.Keymap
}

func (o *ImageOSAdapter) GetKernelArgs() []string {
	return o.OS.KernelArgs
}

func (o *ImageOSAdapter) GetPackages() context.Packages {
	return o.OS.Packages
}

func (o *ImageOSAdapter) GetEnableFIPS() bool {
	return o.OS.EnableFIPS
}

func (o *ImageOSAdapter) GetIsoConfiguration() context.IsoConfiguration {
	return o.OS.IsoConfiguration
}

func (o *ImageOSAdapter) GetRawConfiguration() context.RawConfiguration {
	return o.OS.RawConfiguration
}
