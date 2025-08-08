package combustion

import (
	"github.com/suse-edge/edge-image-builder/pkg/context"
)

func ComponentHelmCharts(ctx *context.Context) ([]context.HelmChart, []context.HelmRepository) {
	if ctx.Definition.GetKubernetes().Version == "" {
		return nil, nil
	}

	const (
		metallbRepositoryName = "suse-edge-metallb"
		metallbNamespace      = "metallb-system"

		endpointCopierOperatorRepositoryName = "suse-edge-endpoint-copier-operator"
		endpointCopierOperatorNamespace      = "endpoint-copier-operator"

		installationNamespace = "kube-system"
	)

	var charts []context.HelmChart
	var repos []context.HelmRepository

	if ctx.Definition.GetKubernetes().Network.APIVIP4 != "" || ctx.Definition.GetKubernetes().Network.APIVIP6 != "" {
		metalLBChart := context.HelmChart{
			Name:                  ctx.ArtifactSources.MetalLB.Chart,
			RepositoryName:        metallbRepositoryName,
			TargetNamespace:       metallbNamespace,
			CreateNamespace:       true,
			InstallationNamespace: installationNamespace,
			Version:               ctx.ArtifactSources.MetalLB.Version,
		}

		endpointCopierOperatorChart := context.HelmChart{
			Name:                  ctx.ArtifactSources.EndpointCopierOperator.Chart,
			RepositoryName:        endpointCopierOperatorRepositoryName,
			TargetNamespace:       endpointCopierOperatorNamespace,
			CreateNamespace:       true,
			InstallationNamespace: installationNamespace,
			Version:               ctx.ArtifactSources.EndpointCopierOperator.Version,
		}

		charts = append(charts, metalLBChart, endpointCopierOperatorChart)

		metallbRepo := context.HelmRepository{
			Name: metallbRepositoryName,
			URL:  ctx.ArtifactSources.MetalLB.Repository,
		}

		endpointCopierOperatorRepo := context.HelmRepository{
			Name: endpointCopierOperatorRepositoryName,
			URL:  ctx.ArtifactSources.EndpointCopierOperator.Repository,
		}

		repos = append(repos, metallbRepo, endpointCopierOperatorRepo)
	}

	return charts, repos
}
