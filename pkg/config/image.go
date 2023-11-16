package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

const (
	ImageTypeISO = "iso"
	ImageTypeRAW = "raw"
)

type ImageConfig struct {
	APIVersion      string          `yaml:"apiVersion"`
	Image           Image           `yaml:"image"`
	ElementalConfig ElementalConfig `yaml:"elementalConfig"`
	OperatingSystem OperatingSystem `yaml:"operatingSystem"`
}

type Image struct {
	ImageType       string `yaml:"imageType"`
	BaseImage       string `yaml:"baseImage"`
	OutputImageName string `yaml:"outputImageName"`
}

type ElementalConfig struct {
	Elemental Elemental `yaml:"elemental"`
}

type Elemental struct {
	Registration struct {
		RegistrationURL string `yaml:"url"`
		CACert          string `yaml:"ca-cert"`
		EmulateTPM      bool   `yaml:"emulate-tpm"`
		EmulateTPMSeed  int    `yaml:"emulated-tpm-seed"`
		AuthType        string `yaml:"auth"`
	} `yaml:"registration"`
}

type OperatingSystem struct {
	KernelArgs []string `yaml:"kernelArgs"`
}

func Parse(data []byte) (*ImageConfig, error) {
	imageConfig := ImageConfig{}

	err := yaml.Unmarshal(data, &imageConfig)
	if err != nil {
		return nil, fmt.Errorf("could not parse the image configuration: %w", err)
	}

	return &imageConfig, nil
}
