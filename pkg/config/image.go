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
	APIVersion string    `yaml:"apiVersion"`
	Image      Image     `yaml:"image"`
	Elemental  Elemental `yaml:"elemental"`
}

type Image struct {
	ImageType       string `yaml:"imageType"`
	BaseImage       string `yaml:"baseImage"`
	OutputImageName string `yaml:"outputImageName"`
}

type Elemental struct {
	Registration struct {
		RegistrationURL string `yaml:"url"`
		CACert          string `yaml:"ca-cert"`
		EmulateTPM      string `yaml:"emulate-tpm"`
		EmulateTPMSeed  string `yaml:"emulated-tpm-seed"`
		AuthType        string `yaml:"auth"`
	} `yaml:"registration"`
}

func Parse(data []byte) (*ImageConfig, error) {
	imageConfig := ImageConfig{}

	err := yaml.Unmarshal(data, &imageConfig)
	if err != nil {
		return nil, fmt.Errorf("could not parse the image configuration: %w", err)
	}

	return &imageConfig, nil
}
