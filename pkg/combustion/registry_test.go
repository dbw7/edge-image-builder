package combustion

import (
	"github.com/suse-edge/edge-image-builder/pkg/context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/image"
)

func TestWriteRegistryScript(t *testing.T) {
	// Setup
	ctx, _, teardown := setupContext(t)
	defer teardown()

	// Test
	_, err := writeRegistryScript(ctx)

	// Verify
	require.NoError(t, err)

	registryScriptPath := filepath.Join(ctx.CombustionDir, registryScriptName)

	foundBytes, err := os.ReadFile(registryScriptPath)
	require.NoError(t, err)

	found := string(foundBytes)
	assert.Contains(t, found, "cp $ARTEFACTS_DIR/registry/hauler /opt/hauler/hauler")
	assert.Contains(t, found, "cp $ARTEFACTS_DIR/registry/*-registry.tar.zst /opt/hauler/")
	assert.Contains(t, found, "systemctl enable eib-embedded-registry.service")
	assert.Contains(t, found, "ExecStartPre=/bin/bash -c \"for file in /opt/hauler/*-registry.tar.zst; do [ -f \\\"\\$file\\\" ] && /opt/hauler/hauler store load -f \\\"\\$file\\\" --tempdir /opt/hauler; done\"\n")
	assert.Contains(t, found, "ExecStart=/opt/hauler/hauler store serve registry -p 6545")
}

func TestIsEmbeddedArtifactRegistryConfigured(t *testing.T) {
	tests := []struct {
		name         string
		ctx          *context.Context
		isConfigured bool
	}{
		{
			name: "Everything Defined",
			ctx: &context.Context{
				Definition: &image.ImageDefinitionAdapter{
					&image.Definition{
						EmbeddedArtifactRegistry: context.EmbeddedArtifactRegistry{
							ContainerImages: []context.ContainerImage{
								{
									Name: "nginx",
								},
							},
						},
						Kubernetes: context.Kubernetes{
							Manifests: context.Manifests{
								URLs: []string{
									"https://k8s.io/examples/application/nginx-app.yaml",
								},
							},
							Helm: context.Helm{
								Charts: []context.HelmChart{
									{
										Name:           "apache",
										RepositoryName: "apache-repo",
										Version:        "10.7.0",
									},
								},
							},
						},
					},
				},
			},
			isConfigured: true,
		},
		{
			name: "Image Defined",
			ctx: &context.Context{
				Definition: &image.ImageDefinitionAdapter{
					&image.Definition{
						EmbeddedArtifactRegistry: context.EmbeddedArtifactRegistry{
							ContainerImages: []context.ContainerImage{
								{
									Name: "nginx",
								},
							},
						},
					},
				},
			},
			isConfigured: true,
		},
		{
			name: "Manifest URL Defined",
			ctx: &context.Context{
				Definition: &image.ImageDefinitionAdapter{
					&image.Definition{
						Kubernetes: context.Kubernetes{
							Manifests: context.Manifests{
								URLs: []string{
									"https://k8s.io/examples/application/nginx-app.yaml",
								},
							},
						},
					},
				},
			},
			isConfigured: true,
		},
		{
			name: "Helm Charts Defined",
			ctx: &context.Context{
				Definition: &image.ImageDefinitionAdapter{
					&image.Definition{
						Kubernetes: context.Kubernetes{
							Helm: context.Helm{
								Charts: []context.HelmChart{
									{
										Name:           "apache",
										RepositoryName: "apache-repo",
										Version:        "10.7.0",
									},
								},
							},
						},
					},
				},
			},
			isConfigured: true,
		},
		{
			name: "None Defined",
			ctx: &context.Context{
				Definition: &image.ImageDefinitionAdapter{
					&image.Definition{
						EmbeddedArtifactRegistry: context.EmbeddedArtifactRegistry{},
						Kubernetes:               context.Kubernetes{},
					},
				},
			},
			isConfigured: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := IsEmbeddedArtifactRegistryConfigured(test.ctx)
			assert.Equal(t, test.isConfigured, result)
		})
	}
}

func TestWriteRegistryMirrorsValid(t *testing.T) {
	// Setup
	ctx, _, teardown := setupContext(t)
	defer teardown()

	hostnames := []string{"hello-world:latest", "rgcrprod.azurecr.us/longhornio/longhorn-ui:v1.5.1", "quay.io"}

	// Test
	err := writeRegistryMirrors(ctx, hostnames)

	// Verify
	require.NoError(t, err)

	manifestFileName := filepath.Join(ctx.ArtefactsDir, k8sDir, registryMirrorsFileName)

	foundBytes, err := os.ReadFile(manifestFileName)
	require.NoError(t, err)

	found := string(foundBytes)
	assert.Contains(t, found, "- \"http://localhost:6545\"")
	assert.Contains(t, found, "docker.io")
	assert.Contains(t, found, "rgcrprod.azurecr.us")
	assert.Contains(t, found, "quay.io")
}

func TestGetImageHostnames(t *testing.T) {
	// Setup
	images := []string{
		"hello-world:latest",
		"quay.io/podman/hello",
		"rgcrprod.azurecr.us/longhornio/longhorn-ui:v1.5.1",
	}
	expectedHostnames := []string{"quay.io", "rgcrprod.azurecr.us"}

	// Test
	hostnames := getImageHostnames(images)

	// Verify
	assert.Equal(t, expectedHostnames, hostnames)
}
