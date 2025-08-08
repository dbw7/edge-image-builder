package validation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"github.com/suse-edge/edge-image-builder/pkg/image"
)

func TestValidateDefinition(t *testing.T) {
	configDir, err := os.MkdirTemp("", "eib-config-")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(configDir)
	}()

	testImagesDir := filepath.Join(configDir, "base-images")
	err = os.MkdirAll(testImagesDir, os.ModePerm)
	require.NoError(t, err)

	fakeBaseImageName := "fake-base.iso"
	_, err = os.Create(filepath.Join(testImagesDir, fakeBaseImageName))
	require.NoError(t, err)

	tests := map[string]struct {
		Definition image.Definition
		Expected   map[string][]string
	}{
		`minimal valid`: {
			Definition: image.Definition{
				APIVersion: "1.2",
				Image: context.Image{
					ImageType:       "iso",
					Arch:            context.ArchTypeX86,
					BaseImage:       fakeBaseImageName,
					OutputImageName: "output.iso",
				},
			},
		},
		`invalid in each`: {
			Definition: image.Definition{
				APIVersion: "1.2",
				Image: context.Image{
					Arch:            context.ArchTypeX86,
					BaseImage:       fakeBaseImageName,
					OutputImageName: "output.iso",
				},
				OperatingSystem: image.OperatingSystem{
					KernelArgs: []string{"foo=", "fips=1"},
				},
				EmbeddedArtifactRegistry: context.EmbeddedArtifactRegistry{
					ContainerImages: []context.ContainerImage{
						{
							Name: "", // trips the missing name validation
						},
					},
				},
				Kubernetes: context.Kubernetes{
					Network: context.Network{},
					Nodes: []context.Node{
						{
							Hostname: "host1",
							Type:     context.KubernetesNodeTypeServer,
						},
						{
							Hostname: "host2",
							Type:     context.KubernetesNodeTypeAgent,
						},
					},
				},
			},
			Expected: map[string][]string{
				imageComponent: {
					"The 'imageType' field is required in the 'image' section.",
				},
				osComponent: {
					"Kernel arguments must be specified as 'key=value'.",
					"FIPS mode has been specified via kernel arguments, please use the 'enableFIPS: true' option instead.",
				},
				registryComponent: {
					"The 'name' field is required for each entry in 'images'.",
				},
				k8sComponent: {
					"The 'apiVIP' field is required in the 'network' section when defining entries under 'nodes'.",
					"The 'apiHost' field is required in the 'network' section when defining entries under 'nodes'.",
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			def := test.Definition
			ctx := context.Context{
				Definition:     &def,
				ImageConfigDir: configDir,
			}
			failures := ValidateDefinition(&ctx)

			for foundComponent, foundComponentFailures := range failures {
				assert.Contains(t, test.Expected, foundComponent)
				assert.Len(t, foundComponentFailures, len(test.Expected[foundComponent]))

				var foundMessages []string
				for _, foundValidation := range foundComponentFailures {
					foundMessages = append(foundMessages, foundValidation.UserMessage)
				}

				for _, expectedMessage := range test.Expected[foundComponent] {
					assert.Contains(t, foundMessages, expectedMessage)
				}
			}
		})
	}
}
