package kubernetes

import (
	config "github.com/suse-edge/edge-image-builder/pkg/config"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRKE2InstallerArtefacts(t *testing.T) {
	x86Artefacts := []string{"rke2.linux-amd64.tar.gz", "sha256sum-amd64.txt"}
	assert.Equal(t, x86Artefacts, rke2InstallerArtefacts(config.ArchTypeX86))

	armArtefacts := []string{"rke2.linux-arm64.tar.gz", "sha256sum-arm64.txt"}
	assert.Equal(t, armArtefacts, rke2InstallerArtefacts(config.ArchTypeARM))
}

func TestRKE2ImageArtefacts(t *testing.T) {
	tests := []struct {
		name              string
		cni               string
		multusEnabled     bool
		arch              config.Arch
		expectedArtefacts []string
		expectedError     string
	}{
		{
			name:          "CNI not specified",
			arch:          config.ArchTypeX86,
			expectedError: "CNI not specified",
		},
		{
			name:          "CNI not supported",
			cni:           "flannel",
			arch:          config.ArchTypeX86,
			expectedError: "unsupported CNI: flannel",
		},
		{
			name: "x86_64 artefacts without CNI",
			cni:  config.CNITypeNone,
			arch: config.ArchTypeX86,
			expectedArtefacts: []string{
				"rke2-images-core.linux-amd64.tar.zst",
			},
		},
		{
			name: "x86_64 artefacts with canal CNI",
			cni:  config.CNITypeCanal,
			arch: config.ArchTypeX86,
			expectedArtefacts: []string{
				"rke2-images-core.linux-amd64.tar.zst",
				"rke2-images-canal.linux-amd64.tar.zst",
			},
		},
		{
			name: "x86_64 artefacts with calico CNI",
			cni:  config.CNITypeCalico,
			arch: config.ArchTypeX86,
			expectedArtefacts: []string{
				"rke2-images-core.linux-amd64.tar.zst",
				"rke2-images-calico.linux-amd64.tar.zst",
			},
		},
		{
			name: "x86_64 artefacts with cilium CNI",
			cni:  config.CNITypeCilium,
			arch: config.ArchTypeX86,
			expectedArtefacts: []string{
				"rke2-images-core.linux-amd64.tar.zst",
				"rke2-images-cilium.linux-amd64.tar.zst",
			},
		},
		{
			name:          "x86_64 artefacts with cilium CNI + multus",
			cni:           config.CNITypeCilium,
			multusEnabled: true,
			arch:          config.ArchTypeX86,
			expectedArtefacts: []string{
				"rke2-images-core.linux-amd64.tar.zst",
				"rke2-images-cilium.linux-amd64.tar.zst",
				"rke2-images-multus.linux-amd64.tar.zst",
			},
		},
		{
			name: "aarch64 artefacts for CNI none",
			cni:  config.CNITypeNone,
			arch: config.ArchTypeARM,
			expectedArtefacts: []string{
				"rke2-images-core.linux-arm64.tar.zst",
			},
		},
		{
			name: "aarch64 artefacts with canal CNI",
			cni:  config.CNITypeCanal,
			arch: config.ArchTypeARM,
			expectedArtefacts: []string{
				"rke2-images-core.linux-arm64.tar.zst",
				"rke2-images-canal.linux-arm64.tar.zst",
			},
		},
		{
			name: "aarch64 artefacts with calico CNI",
			cni:  config.CNITypeCalico,
			arch: config.ArchTypeARM,
			expectedArtefacts: []string{
				"rke2-images-core.linux-arm64.tar.zst",
				"rke2-images-calico.linux-arm64.tar.zst",
			},
		},
		{
			name: "aarch64 artefacts with cilium CNI",
			cni:  config.CNITypeCilium,
			arch: config.ArchTypeARM,
			expectedArtefacts: []string{
				"rke2-images-core.linux-arm64.tar.zst",
				"rke2-images-cilium.linux-arm64.tar.zst",
			},
		},
		{
			name:          "aarch64 artefacts with canal CNI + multus",
			cni:           config.CNITypeCanal,
			multusEnabled: true,
			arch:          config.ArchTypeARM,
			expectedArtefacts: []string{
				"rke2-images-core.linux-arm64.tar.zst",
				"rke2-images-canal.linux-arm64.tar.zst",
				"rke2-images-multus.linux-arm64.tar.zst",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			artefacts, err := rke2ImageArtefacts(test.cni, test.multusEnabled, test.arch)

			if test.expectedError != "" {
				require.EqualError(t, err, test.expectedError)
				assert.Nil(t, artefacts)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.expectedArtefacts, artefacts)
			}
		})
	}
}

func TestK3sInstallerArtefacts(t *testing.T) {
	x86Artefacts := []string{"k3s"}
	assert.Equal(t, x86Artefacts, k3sInstallerArtefacts(config.ArchTypeX86))

	armArtefacts := []string{"k3s-arm64"}
	assert.Equal(t, armArtefacts, k3sInstallerArtefacts(config.ArchTypeARM))
}

func TestK3sImageArtefacts(t *testing.T) {
	x86Artefacts := []string{"k3s-airgap-images-amd64.tar.zst"}
	assert.Equal(t, x86Artefacts, k3sImageArtefacts(config.ArchTypeX86))

	armArtefacts := []string{"k3s-airgap-images-arm64.tar.zst"}
	assert.Equal(t, armArtefacts, k3sImageArtefacts(config.ArchTypeARM))
}
