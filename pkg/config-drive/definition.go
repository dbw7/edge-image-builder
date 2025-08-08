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

func ParseConfigDriveDefinition(data []byte) (context.Definition, error) {
	var definition Definition

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)

	if err := decoder.Decode(&definition); err != nil {
		return nil, fmt.Errorf("could not parse the config drive definition: %w", err)
	}

	if !version.IsSchemaVersionSupported(definition.APIVersion) {
		return nil, context.ErrorInvalidSchemaVersion
	}

	return &ConfigDriveDefinitionAdapter{Definition: &definition}, nil
}

type ConfigDriveDefinitionAdapter struct {
	*Definition
}

func (a *ConfigDriveDefinitionAdapter) GetAPIVersion() string {
	return a.APIVersion
}

func (a *ConfigDriveDefinitionAdapter) GetImage() context.Image {
	return context.Image{}
}

func (a *ConfigDriveDefinitionAdapter) GetOperatingSystem() context.OperatingSystemInterface {
	return &ConfigDriveOSAdapter{OS: &a.OperatingSystem}
}

func (a *ConfigDriveDefinitionAdapter) GetKubernetes() context.Kubernetes {
	return a.Kubernetes
}

func (a *ConfigDriveDefinitionAdapter) GetEmbeddedArtifactRegistry() context.EmbeddedArtifactRegistry {
	return a.EmbeddedArtifactRegistry
}

type ConfigDriveOSAdapter struct {
	OS *OperatingSystem
}

func (o *ConfigDriveOSAdapter) GetUsers() []context.OperatingSystemUser {
	return o.OS.Users
}

func (o *ConfigDriveOSAdapter) GetGroups() []context.OperatingSystemGroup {
	return o.OS.Groups
}

func (o *ConfigDriveOSAdapter) GetSystemd() context.Systemd {
	return o.OS.Systemd
}

func (o *ConfigDriveOSAdapter) GetSuma() context.Suma {
	return o.OS.Suma
}

func (o *ConfigDriveOSAdapter) GetTime() context.Time {
	return o.OS.Time
}

func (o *ConfigDriveOSAdapter) GetProxy() context.Proxy {
	return o.OS.Proxy
}

func (o *ConfigDriveOSAdapter) GetKeymap() string {
	return o.OS.Keymap
}

func (o *ConfigDriveOSAdapter) GetKernelArgs() []string {
	return []string{}
}

func (o *ConfigDriveOSAdapter) GetPackages() context.Packages {
	return context.Packages{}
}

func (o *ConfigDriveOSAdapter) GetEnableFIPS() bool {
	return false
}

func (o *ConfigDriveOSAdapter) GetIsoConfiguration() context.IsoConfiguration {
	return context.IsoConfiguration{}
}

func (o *ConfigDriveOSAdapter) GetRawConfiguration() context.RawConfiguration {
	return context.RawConfiguration{}
}
