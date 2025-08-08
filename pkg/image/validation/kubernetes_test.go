package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/suse-edge/edge-image-builder/pkg/config"
	"github.com/suse-edge/edge-image-builder/pkg/image"
	"gopkg.in/yaml.v3"
)

var validNetwork = config.Network{
	APIHost: "host.com",
	APIVIP4: "192.168.100.1",
}

func TestValidateKubernetes(t *testing.T) {
	configDir, err := os.MkdirTemp("", "eib-config-")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(configDir))
	}()

	valuesDir := filepath.Join(configDir, "kubernetes", "helm", "values")
	require.NoError(t, os.MkdirAll(valuesDir, os.ModePerm))

	apacheValuesPath := filepath.Join(valuesDir, "apache-values.yaml")
	require.NoError(t, os.WriteFile(apacheValuesPath, []byte(""), 0o600))

	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`not defined`: {
			K8s: config.Kubernetes{},
		},
		`all valid`: {
			K8s: config.Kubernetes{
				Version: "v1.30.3+k3s1",
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname:    "server",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
					{
						Hostname: "agent1",
						Type:     config.KubernetesNodeTypeAgent,
					},
				},
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:                  "apache",
							RepositoryName:        "apache-repo",
							TargetNamespace:       "web",
							CreateNamespace:       true,
							InstallationNamespace: "kube-system",
							Version:               "10.7.0",
							ValuesFile:            "apache-values.yaml",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
		},
		`failures all sections`: {
			K8s: config.Kubernetes{
				Version: "v1.30.3",
				Network: config.Network{
					APIHost: "host.com",
					APIVIP4: "127.0.0.1",
					APIVIP6: "ff02::1",
				},
				Nodes: []config.Node{
					{
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
					{
						Hostname: "valid",
						Type:     config.KubernetesNodeTypeAgent,
					},
				},
				Manifests: config.Manifests{
					URLs: []string{
						"example.com",
					},
				},
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "",
							RepositoryName: "another-apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"The 'hostname' field is required for entries in the 'nodes' section.",
				"Entries in 'urls' must begin with either 'http://' or 'https://'.",
				"Helm chart 'name' field must be defined.",
				"Helm repository 'name' field for \"apache-repo\" must match the 'repositoryName' field in at least one defined Helm chart.",
				"Helm chart 'repositoryName' \"another-apache-repo\" for Helm chart \"\" does not match the name of any defined repository.",
				"Non-unicast cluster API address (127.0.0.1) for field 'apiVIP' is invalid.",
				"Non-unicast cluster API address (ff02::1) for field 'apiVIP6' is invalid.",
				fmt.Sprintf("Kubernetes server config could not be found at '%s'; dual-stack configuration requires a valid cluster-cidr and service-cidr.", filepath.Join(configDir, "kubernetes", "config", "server.yaml")),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := config.Context{
				ImageConfigDir: configDir,
				Definition: &image.Definition{
					Kubernetes: test.K8s,
				},
			}
			failures := validateKubernetes(&ctx)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}

func TestIsKubernetesDefined(t *testing.T) {
	result := isKubernetesDefined(&config.Kubernetes{
		Version: "v1.30.3+k3s1",
	})
	assert.True(t, result)

	result = isKubernetesDefined(&config.Kubernetes{
		Network:   config.Network{},
		Nodes:     []config.Node{},
		Manifests: config.Manifests{},
		Helm:      config.Helm{},
	})
	assert.False(t, result)
}

func TestValidateNodes(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`valid`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "agent1",
						Type:     config.KubernetesNodeTypeAgent,
					},
					{
						Hostname:    "server",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
				},
			},
		},
		`no nodes`: {
			K8s: config.Kubernetes{
				Nodes: []config.Node{},
			},
		},
		`no hostname`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "host1",
						Type:     config.KubernetesNodeTypeServer,
					},
					{
						Type: config.KubernetesNodeTypeServer,
					},
				},
			},
			ExpectedFailedMessages: []string{
				"The 'hostname' field is required for entries in the 'nodes' section.",
			},
		},
		`missing type`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "host1",
						Type:     config.KubernetesNodeTypeServer,
					},
					{
						Hostname: "valid",
					},
				},
			},
			ExpectedFailedMessages: []string{
				fmt.Sprintf("The 'type' field for entries in the 'nodes' section must be one of: %s", strings.Join(validNodeTypes, ", ")),
			},
		},
		`invalid type`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "valid",
						Type:     config.KubernetesNodeTypeServer,
					},
					{
						Hostname: "invalid",
						Type:     "abnormal",
					},
				},
			},
			ExpectedFailedMessages: []string{
				fmt.Sprintf("The 'type' field for entries in the 'nodes' section must be one of: %s", strings.Join(validNodeTypes, ", ")),
			},
		},
		`incorrect initialiser type`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "valid",
						Type:     config.KubernetesNodeTypeServer,
					},
					{
						Hostname:    "invalid",
						Initialiser: true,
						Type:        config.KubernetesNodeTypeAgent,
					},
				},
			},
			ExpectedFailedMessages: []string{
				fmt.Sprintf("The node labeled with 'initialiser' must be of type '%s'.", config.KubernetesNodeTypeServer),
			},
		},
		`duplicate entries`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname:    "foo",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
					{
						Hostname: "bar",
						Type:     config.KubernetesNodeTypeAgent,
					},
					{
						Hostname: "bar",
						Type:     config.KubernetesNodeTypeAgent,
					},
					{
						Hostname: "foo",
						Type:     config.KubernetesNodeTypeAgent,
					},
				},
			},
			ExpectedFailedMessages: []string{
				"The 'nodes' section contains duplicate entries: bar, foo",
			},
		},
		`no server node`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname: "foo",
						Type:     config.KubernetesNodeTypeAgent,
					},
					{
						Hostname: "bar",
						Type:     config.KubernetesNodeTypeAgent,
					},
				},
			},
			ExpectedFailedMessages: []string{
				fmt.Sprintf("There must be at least one node of type '%s' defined.", config.KubernetesNodeTypeServer),
			},
		},
		`multiple initialisers`: {
			K8s: config.Kubernetes{
				Network: validNetwork,
				Nodes: []config.Node{
					{
						Hostname:    "foo",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
					{
						Hostname:    "bar",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Only one node may be specified as the cluster initializer.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			failures := validateNodes(&k)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}

func TestValidateManifestURLs(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`valid`: {
			K8s: config.Kubernetes{
				Manifests: config.Manifests{
					URLs: []string{
						"http://valid1.com",
						"https://valid2.com",
					},
				},
			},
		},
		`no URLs`: {
			K8s: config.Kubernetes{
				Manifests: config.Manifests{},
			},
		},
		`invalid prefix`: {
			K8s: config.Kubernetes{
				Manifests: config.Manifests{
					URLs: []string{
						"http://valid.com",
						"https://also-valid.com",
						"invalid.com",
						"nope.com",
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Entries in 'urls' must begin with either 'http://' or 'https://'.",
				"Entries in 'urls' must begin with either 'http://' or 'https://'.",
			},
		},
		`duplicate URLs`: {
			K8s: config.Kubernetes{
				Manifests: config.Manifests{
					URLs: []string{
						"http://foo.com",
						"http://bar.com",
						"http://foo.com",
						"http://bar.com",
					},
				},
			},
			ExpectedFailedMessages: []string{
				"The 'urls' field contains duplicate entries: http://foo.com",
				"The 'urls' field contains duplicate entries: http://bar.com",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			failures := validateManifestURLs(&k)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}
		})
	}
}

func TestValidateHelmCharts(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`valid`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:                  "apache",
							RepositoryName:        "apache-repo",
							TargetNamespace:       "web",
							CreateNamespace:       true,
							InstallationNamespace: "kube-system",
							Version:               "10.7.0",
						},
						{
							Name:                  "apache",
							ReleaseName:           "apache2",
							RepositoryName:        "apache-repo",
							TargetNamespace:       "web",
							CreateNamespace:       true,
							InstallationNamespace: "kube-system",
							Version:               "10.7.0",
						},
						{
							Name:                  "apache",
							ReleaseName:           "apache3",
							RepositoryName:        "apache-repo",
							TargetNamespace:       "web2",
							CreateNamespace:       true,
							InstallationNamespace: "kube-system",
							Version:               "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
							Authentication: config.HelmAuthentication{
								Username: "user",
								Password: "pass",
							},
						},
					},
				},
			},
		},
		`helm no repos`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm charts defined with no Helm repositories defined.",
			},
		},
		`helm chart no name`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'name' field must be defined.",
			},
		},
		`helm chart undefined repository name`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "kubevirt",
							RepositoryName: "suse-edge",
							Version:        "0.2.2",
						},
						{
							Name:           "metallb",
							RepositoryName: "",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "suse-edge",
							URL:  "https://suse-edge.github.io/charts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'repositoryName' field for \"metallb\" must be defined.",
			},
		},
		`helm chart no matching repository name`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "kubevirt",
							RepositoryName: "suse-edge",
							Version:        "0.2.2",
						},
						{
							Name:           "metallb",
							RepositoryName: "this-is-not-suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "suse-edge",
							URL:  "https://suse-edge.github.io/charts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'repositoryName' \"this-is-not-suse-edge\" for Helm chart \"metallb\" does not match the name of any defined repository.",
			},
		},
		`helm chart no version`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'version' field for \"apache\" field must be defined.",
			},
		},
		`helm chart create namespace no target`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:            "apache",
							RepositoryName:  "apache-repo",
							Version:         "10.7.0",
							CreateNamespace: true,
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'createNamespace' field for \"apache\" cannot be true without 'targetNamespace' being defined.",
			},
		},
		`helm chart duplicate name no release name`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm charts with the same 'name' require a unique 'releaseName'. Duplicate found:\nName: 'apache', Release name: ''",
			},
		},
		// This configuration would be valid for a regular chart deployment;
		// however, since we use the Helm controller, a different target namespace is not sufficient. The release names must be different.
		`helm chart duplicate name, same release name, different target namespaces`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:            "apache",
							ReleaseName:     "apache-deployment",
							TargetNamespace: "web",
							RepositoryName:  "apache-repo",
							Version:         "10.7.0",
						},
						{
							Name:            "apache",
							ReleaseName:     "apache-deployment",
							TargetNamespace: "web2",
							RepositoryName:  "apache-repo",
							Version:         "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm charts with the same 'name' require a unique 'releaseName'. Duplicate found:\nName: 'apache', Release name: 'apache-deployment'",
			},
		},
		`helm chart invalid values file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
							ValuesFile:     "invalid",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'valuesFile' field for \"apache\" must be the name of a valid yaml file ending in '.yaml' or '.yml'.",
			},
		},
		`helm chart nonexistent values file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
							ValuesFile:     "nonexistent.yaml",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart values file 'nonexistent.yaml' could not be found at 'kubernetes/helm/values/nonexistent.yaml'.",
			},
		},
		`helm repository no name`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "",
							URL:  "https://suse-edge.github.io/charts",
						},
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'name' field must be defined.",
			},
		},
		`helm repository no url`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'url' field for \"apache-repo\" must be defined.",
			},
		},
		`helm repository invalid url`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "invalid.repo.io/bitnami",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'url' field for \"apache-repo\" must begin with either 'oci://', 'http://', or 'https://'.",
			},
		},
		`helm repository username no password`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
							Authentication: config.HelmAuthentication{
								Username: "user",
								Password: "",
							},
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'password' field not defined for \"apache-repo\".",
			},
		},
		`helm repository password no username`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
							Authentication: config.HelmAuthentication{
								Username: "",
								Password: "pass",
							},
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'username' field not defined for \"apache-repo\".",
			},
		},
		`helm repository both skipTLSVerify and plainHTTP true`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:          "apache-repo",
							URL:           "oci://registry-1.docker.io/bitnamicharts",
							SkipTLSVerify: true,
							PlainHTTP:     true,
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'plainHTTP' and 'skipTLSVerify' fields for \"apache-repo\" cannot both be true.",
			},
		},
		`helm repository skipTLSVerify true for http`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "metallb",
							RepositoryName: "suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:          "suse-edge",
							URL:           "http://suse-edge.github.io/charts",
							SkipTLSVerify: true,
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'url' field for \"suse-edge\" contains 'http://' but 'plainHTTP' field is false.",
				"Helm repository 'url' field for \"suse-edge\" contains 'http://' but 'skipTLSVerify' field is true.",
			},
		},
		`helm repository plainHTTP false for http`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "metallb",
							RepositoryName: "suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:      "suse-edge",
							URL:       "http://suse-edge.github.io/charts",
							PlainHTTP: false,
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'url' field for \"suse-edge\" contains 'http://' but 'plainHTTP' field is false.",
			},
		},
		`helm repository plainHTTP true for https`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "metallb",
							RepositoryName: "suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:      "suse-edge",
							URL:       "https://suse-edge.github.io/charts",
							PlainHTTP: true,
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'url' field for \"suse-edge\" contains 'https://' but 'plainHTTP' field is true.",
			},
		},
		`helm repository plainHTTP and ca file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "metallb",
							RepositoryName: "suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:      "suse-edge",
							URL:       "http://suse-edge.github.io/charts",
							PlainHTTP: true,
							CAFile:    "suse-edge.crt",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'caFile' field for \"suse-edge\" cannot be defined while 'plainHTTP' is true.",
				"Helm repository 'url' field for \"suse-edge\" contains 'http://' but 'caFile' field is defined.",
				"Helm repo cert file/bundle 'suse-edge.crt' could not be found at 'kubernetes/helm/certs/suse-edge.crt'.",
			},
		},
		`helm repository skipTLSVerify and ca file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "metallb",
							RepositoryName: "suse-edge",
							Version:        "0.14.3",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:          "suse-edge",
							URL:           "https://suse-edge.github.io/charts",
							SkipTLSVerify: true,
							CAFile:        "suse-edge.crt",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repository 'caFile' field for \"suse-edge\" cannot be defined while 'skipTLSVerify' is true.",
				"Helm repo cert file/bundle 'suse-edge.crt' could not be found at 'kubernetes/helm/certs/suse-edge.crt'.",
			},
		},
		`helm repo nonexistent cert file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:   "apache-repo",
							URL:    "oci://registry-1.docker.io/bitnamicharts",
							CAFile: "nonexistent-apache.crt",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm repo cert file/bundle 'nonexistent-apache.crt' could not be found at 'kubernetes/helm/certs/nonexistent-apache.crt'.",
			},
		},
		`helm repo invalid cert file`: {
			K8s: config.Kubernetes{
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "apache",
							RepositoryName: "apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name:   "apache-repo",
							URL:    "oci://registry-1.docker.io/bitnamicharts",
							CAFile: "invalid-cert",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Helm chart 'caFile' field for \"apache-repo\" must be the name of a valid cert file/bundle with one of the " +
					"following extensions: .pem, .crt, .cer",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			failures := validateHelm(&k, "kubernetes/helm/values", "kubernetes/helm/certs")
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}
		})
	}
}

func TestValidateAdditionalArtifacts(t *testing.T) {
	configDir, err := os.MkdirTemp("", "eib-config-")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(configDir))
	}()

	manifestsDir := filepath.Join(configDir, "kubernetes", "manifests")
	require.NoError(t, os.MkdirAll(manifestsDir, os.ModePerm))

	testManifest := filepath.Join(manifestsDir, "manifest.yaml")
	require.NoError(t, os.WriteFile(testManifest, []byte(""), 0o600))

	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`missing versions all sections`: {
			K8s: config.Kubernetes{
				Manifests: config.Manifests{
					URLs: []string{
						"example.com",
					},
				},
				Helm: config.Helm{
					Charts: []config.HelmChart{
						{
							Name:           "",
							RepositoryName: "another-apache-repo",
							Version:        "10.7.0",
						},
					},
					Repositories: []config.HelmRepository{
						{
							Name: "apache-repo",
							URL:  "oci://registry-1.docker.io/bitnamicharts",
						},
					},
				},
			},
			ExpectedFailedMessages: []string{
				"Kubernetes version must be defined when Helm charts are specified",
				"Kubernetes version must be defined when manifest URLs are specified",
				"Kubernetes version must be defined when local manifests are configured",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := &config.Context{
				ImageConfigDir: configDir,
				Definition: &image.Definition{
					Kubernetes: test.K8s,
				},
			}
			failures := validateAdditionalArtifacts(ctx)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}

func TestValidateNetwork(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ExpectedFailedMessages []string
	}{
		`no network defined, no nodes defined`: {
			K8s: config.Kubernetes{
				Network: config.Network{},
			},
		},
		`IPv6 in apiVIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "fd12:3456:789a::21",
				},
			},
			ExpectedFailedMessages: []string{
				"Only IPv4 addresses are valid for field 'apiVIP'.",
			},
		},
		`IPv4 in apiVIP6`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "192.168.1.1",
				},
			},
			ExpectedFailedMessages: []string{
				"Only IPv6 addresses are valid for field 'apiVIP6'.",
			},
		},
		`no network defined, nodes defined`: {
			K8s: config.Kubernetes{
				Network: config.Network{},
				Nodes: []config.Node{
					{
						Hostname:    "node1",
						Type:        "server",
						Initialiser: false,
					},
					{
						Hostname:    "node2",
						Type:        "server",
						Initialiser: false,
					},
				},
			},
			ExpectedFailedMessages: []string{
				"At least one of the (`apiVIP`, `apiVIP6`) fields is required in the 'network' section for multi node clusters.",
			},
		},
		`valid IPv4`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
				},
			},
		},
		`invalid IPv4`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "500.168.1.1",
				},
			},
			ExpectedFailedMessages: []string{
				"Invalid address value \"500.168.1.1\" for field 'apiVIP'.",
			},
		},
		`valid IPv6`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "fd12:3456:789a::21",
				},
			},
		},
		`invalid IPv6`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "xxxx:3456:789a::21",
				},
			},
			ExpectedFailedMessages: []string{
				"Invalid address value \"xxxx:3456:789a::21\" for field 'apiVIP6'.",
			},
		},
		`valid dualstack`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
		},
		`invalid dualstack IPv4 non unicast`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "127.0.0.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (127.0.0.1) for field 'apiVIP' is invalid.",
			},
		},
		`invalid dualstack IPv6 non unicast`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "ff02::1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (ff02::1) for field 'apiVIP6' is invalid.",
			},
		},
		`invalid dualstack both non unicast`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "127.0.0.1",
					APIVIP6: "ff02::1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (127.0.0.1) for field 'apiVIP' is invalid.",
				"Non-unicast cluster API address (ff02::1) for field 'apiVIP6' is invalid.",
			},
		},
		`invalid dualstack IPv4 not valid`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "500.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ExpectedFailedMessages: []string{
				"Invalid address value \"500.168.1.1\" for field 'apiVIP'.",
			},
		},
		`invalid dualstack IPv6 not valid`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "xxxx:3456:789a::21",
				},
			},
			ExpectedFailedMessages: []string{
				"Invalid address value \"xxxx:3456:789a::21\" for field 'apiVIP6'.",
			},
		},
		`undefined v4 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "0.0.0.0",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (0.0.0.0) for field 'apiVIP' is invalid.",
			},
		},
		`undefined v6 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "::",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (::) for field 'apiVIP6' is invalid.",
			},
		},
		`loopback v4 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "127.0.0.1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (127.0.0.1) for field 'apiVIP' is invalid.",
			},
		},
		`loopback v6 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "::1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (::1) for field 'apiVIP6' is invalid.",
			},
		},
		`multicast v4 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "224.224.224.224",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (224.224.224.224) for field 'apiVIP' is invalid.",
			},
		},
		`multicast v6 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "FF01::1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (FF01::1) for field 'apiVIP6' is invalid.",
			},
		},
		`link-local v4 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "169.254.1.1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (169.254.1.1) for field 'apiVIP' is invalid.",
			},
		},
		`link-local v6 VIP`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP6: "FE80::1",
				},
			},
			ExpectedFailedMessages: []string{
				"Non-unicast cluster API address (FE80::1) for field 'apiVIP6' is invalid.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			failures := validateNetwork(&k)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}

func TestValidateConfigInvalidServerConfigNotConfigured(t *testing.T) {
	k8s := config.Kubernetes{Network: config.Network{
		APIVIP4: "192.168.1.1",
		APIVIP6: "fd12:3456:789a::21",
	}}

	failures := validateNetworkingConfig(&k8s, "fake-path")

	assert.Len(t, failures, 1)

	assert.Contains(t, failures[0].UserMessage, "Kubernetes server config could not be found at 'fake-path'; dual-stack configuration requires a valid cluster-cidr and service-cidr.")
}

func TestValidateConfigValidAPIVIPNotConfigured(t *testing.T) {
	k8s := config.Kubernetes{}

	failures := validateNetworkingConfig(&k8s, "")
	assert.Len(t, failures, 0)
}

func TestValidateConfigValidIPv4Prio(t *testing.T) {
	k8s := config.Kubernetes{Network: config.Network{
		APIVIP4: "192.168.1.1",
		APIVIP6: "fd12:3456:789a::21",
	}}

	configDir, err := os.MkdirTemp("", "eib-config-")
	require.NoError(t, err)

	defer func() {
		assert.NoError(t, os.RemoveAll(configDir))
	}()

	serverConfigDir := filepath.Join(configDir, "kubernetes", "config")
	require.NoError(t, os.MkdirAll(serverConfigDir, os.ModePerm))

	serverConfig := map[string]any{
		"cluster-cidr": "10.42.0.0/16,fd12:3456:789b::/48",
		"service-cidr": "10.43.0.0/16,fd12:3456:789c::/112",
	}

	b, err := yaml.Marshal(serverConfig)
	require.NoError(t, err)

	configFile := filepath.Join(serverConfigDir, "server.yaml")
	require.NoError(t, os.WriteFile(configFile, b, 0o600))

	failures := validateNetworkingConfig(&k8s, configFile)

	assert.Len(t, failures, 0)
}

func TestValidateCIDRConfig(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ServerConfig           map[string]any
		ExpectedFailedMessages []string
	}{
		`cluster cidr not configured`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"service-cidr": "10.43.0.0/16,fd12:3456:789c::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config must contain a valid cluster-cidr when configuring dual-stack",
			},
		},
		`service cidr not configured`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "fd12:3456:789b::/48,10.42.0.0/16",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config must contain a valid service-cidr when configuring dual-stack",
			},
		},
		`invalid IPv4`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "500.42.0.0/16,fd12:3456:789b::/48",
				"service-cidr": "500.43.0.0/16,fd12:3456:789c::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value '500.42.0.0/16' could not be parsed",
				"Kubernetes server config service-cidr value '500.43.0.0/16' could not be parsed",
			},
		},
		`invalid IPv6`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,xxxx:3456:789b::/48",
				"service-cidr": "10.43.0.0/16,xxxx:3456:789c::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value 'xxxx:3456:789b::/48' could not be parsed",
				"Kubernetes server config service-cidr value 'xxxx:3456:789c::/112' could not be parsed",
			},
		},
		`invalid IPv6 prefix`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,fd12:3456:789a::/480",
				"service-cidr": "10.43.0.0/16,fd12:3456:789a::/1122",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value 'fd12:3456:789a::/480' could not be parsed",
				"Kubernetes server config service-cidr value 'fd12:3456:789a::/1122' could not be parsed",
			},
		},
		`invalid IPv4 prefix`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/50,fd12:3456:789a::/48",
				"service-cidr": "10.43.0.0/50,fd12:3456:789a::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value '10.42.0.0/50' could not be parsed",
				"Kubernetes server config service-cidr value '10.43.0.0/50' could not be parsed",
			},
		},
		`invalid IPv4 non-unicast`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "127.0.0.1/16,fd12:3456:789a::/48",
				"service-cidr": "127.0.0.1/16,fd12:3456:789a::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value '127.0.0.1/16' must be a valid unicast address",
				"Kubernetes server config service-cidr value '127.0.0.1/16' must be a valid unicast address",
			},
		},
		`invalid IPv6 non-unicast`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,FF01::/48",
				"service-cidr": "10.43.0.0/16,FF01::/112",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr value 'FF01::/48' must be a valid unicast address",
				"Kubernetes server config service-cidr value 'FF01::/112' must be a valid unicast address",
			},
		},
		`mismatched prio`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,fd12:3456:789b::/48",
				"service-cidr": "fd12:3456:789c::/112,10.43.0.0/16",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config cluster-cidr cannot prioritize one address family while service-cidr prioritizes another; both must have the same priority",
			},
		},
		`single cidr IPv6`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"cluster-cidr": "fd12:3456:789b::/48",
				"service-cidr": "fd12:3456:789c::/112",
			},
		},
		`single cidr IPv4`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16",
				"service-cidr": "10.43.0.0/16",
			},
		},
		`valid dualstack IPv6 prio`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "fd12:3456:789b::/48,10.42.0.0/16",
				"service-cidr": "fd12:3456:789c::/112,10.43.0.0/16",
			},
		},
		`valid dualstack IPv4 prio`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,fd12:3456:789b::/48",
				"service-cidr": "10.43.0.0/16,fd12:3456:789c::/112",
			},
		},
		`more than 2 cidrs`: {
			K8s: config.Kubernetes{
				Network: config.Network{
					APIVIP4: "192.168.1.1",
					APIVIP6: "fd12:3456:789a::21",
				},
			},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,fd12:3456:789b::/48,10.42.0.0/16",
				"service-cidr": "10.43.0.0/16,fd12:3456:789c::/112,10.43.0.0/16",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config must contain a valid cluster-cidr when configuring dual-stack",
				"Kubernetes server config must contain a valid service-cidr when configuring dual-stack",
			},
		},
		`no APIVIPs`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"cluster-cidr": "10.42.0.0/16,fd12:3456:789b::/48",
				"service-cidr": "10.43.0.0/16,fd12:3456:789c::/112",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			sc := test.ServerConfig
			failures := validateCIDRConfig(&k, sc)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}

func TestValidateNodeIP(t *testing.T) {
	tests := map[string]struct {
		K8s                    config.Kubernetes
		ServerConfig           map[string]any
		ExpectedFailedMessages []string
	}{
		`single IPv4`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0",
			},
		},
		`single IPv6`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "fd12:3456:789a::21",
			},
		},
		`multiple nodes, one server`: {
			K8s: config.Kubernetes{
				Nodes: []config.Node{
					{
						Hostname:    "server1",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
					{
						Hostname: "agent",
						Type:     config.KubernetesNodeTypeAgent,
					},
					{
						Hostname: "agent2",
						Type:     config.KubernetesNodeTypeAgent,
					},
				},
			},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0",
			},
		},
		`multiple nodes, multiple servers`: {
			K8s: config.Kubernetes{
				Nodes: []config.Node{
					{
						Hostname: "server1",
						Type:     config.KubernetesNodeTypeServer,
					},
					{
						Hostname:    "server2",
						Type:        config.KubernetesNodeTypeServer,
						Initialiser: true,
					},
				},
			},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip can not be specified when there is more than one Kubernetes server node",
			},
		},
		`node ip family both same invalid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0,10.43.0.0",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip cannot contain addresses of the same IP address family; one must be IPv4, and the other IPv6",
			},
		},
		`node ip dualstack valid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0,fd12:3456:789a::21",
			},
		},
		`node ip non-unicast IPv4 invalid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "127.0.0.1,fd12:3456:789a::21",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip value '127.0.0.1' must be a valid unicast address",
			},
		},
		`node ip non-unicast IPv6 invalid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "10.42.0.0,FF01::",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip value 'FF01::' must be a valid unicast address",
			},
		},
		`node ip IPv4 invalid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "550.2.2.2",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip value '550.2.2.2' could not be parsed",
			},
		},
		`node ip IPv6 invalid`: {
			K8s: config.Kubernetes{},
			ServerConfig: map[string]any{
				"node-ip": "xxxx:3456:789a::21",
			},
			ExpectedFailedMessages: []string{
				"Kubernetes server config node-ip value 'xxxx:3456:789a::21' could not be parsed",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			k := test.K8s
			sc := test.ServerConfig
			failures := validateNodeIP(&k, sc)
			assert.Len(t, failures, len(test.ExpectedFailedMessages))

			var foundMessages []string
			for _, foundValidation := range failures {
				foundMessages = append(foundMessages, foundValidation.UserMessage)
			}

			for _, expectedMessage := range test.ExpectedFailedMessages {
				assert.Contains(t, foundMessages, expectedMessage)
			}

		})
	}
}
