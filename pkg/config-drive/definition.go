package configdrive

import (
	"bytes"
	"fmt"

	"github.com/suse-edge/edge-image-builder/pkg/context"
	"github.com/suse-edge/edge-image-builder/pkg/version"
	"gopkg.in/yaml.v3"
)

type Definition struct {
	APIVersion               string                           `yaml:"apiVersion"`
	OperatingSystem          OperatingSystem                  `yaml:"operatingSystem"`
	EmbeddedArtifactRegistry context.EmbeddedArtifactRegistry `yaml:"embeddedArtifactRegistry"`
	Kubernetes               context.Kubernetes               `yaml:"kubernetes"`
}

type OperatingSystem struct {
	Groups  []context.OperatingSystemGroup `yaml:"groups"`
	Users   []context.OperatingSystemUser  `yaml:"users"`
	Systemd context.Systemd                `yaml:"systemd"`
	Suma    context.Suma                   `yaml:"suma"`
	Time    context.Time                   `yaml:"time"`
	Proxy   context.Proxy                  `yaml:"proxy"`
	Keymap  string                         `yaml:"keymap"`
}

var _ context.Definition = (*Definition)(nil)

func ParseConfigDriveDefinition(data []byte) (*Definition, error) {
	var definition Definition

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&definition); err != nil {
		return nil, fmt.Errorf("could not parse the config drive definition: %w", err)
	}

	if !version.IsSchemaVersionSupported(definition.APIVersion) {
		return nil, context.ErrorInvalidSchemaVersion
	}

	return &definition, nil
}

func (d *Definition) GetAPIVersion() string {
	return d.APIVersion
}

func (d *Definition) GetImage() context.Image {
	return context.Image{}
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
	return []string{}
}

func (o *OperatingSystem) GetPackages() context.Packages {
	return context.Packages{}
}

func (o *OperatingSystem) GetEnableFIPS() bool {
	return false
}

func (o *OperatingSystem) GetIsoConfiguration() context.IsoConfiguration {
	return context.IsoConfiguration{}
}

func (o *OperatingSystem) GetRawConfiguration() context.RawConfiguration {
	return context.RawConfiguration{}
}
