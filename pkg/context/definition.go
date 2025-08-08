package context

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Constants
const (
	TypeISO = "iso"
	TypeRAW = "raw"

	ArchTypeX86 Arch = "x86_64"
	ArchTypeARM Arch = "aarch64"

	KubernetesDistroRKE2 = "rke2"
	KubernetesDistroK3S  = "k3s"

	KubernetesNodeTypeServer = "server"
	KubernetesNodeTypeAgent  = "agent"

	CNITypeNone   = "none"
	CNITypeCilium = "cilium"
	CNITypeCanal  = "canal"
	CNITypeCalico = "calico"
)

var (
	diskSizeRegexp            = regexp.MustCompile(`^([1-9]\d+|[1-9])+([MGT])`)
	ErrorInvalidSchemaVersion = errors.New("invalid schema version")
)

type Definition interface {
	APIVersion               string
Kubernetes               Kubernetes
EmbeddedArtifactRegistry EmbeddedArtifactRegistry

GetImage() Image
GetOperatingSystem() OperatingSystemInterface
}

type OperatingSystemInterface interface {
	GetUsers() []OperatingSystemUser
	GetGroups() []OperatingSystemGroup
	GetSystemd() Systemd
	GetSuma() Suma
	GetTime() Time
	GetProxy() Proxy
	GetKeymap() string
	GetKernelArgs() []string
	GetPackages() Packages
	GetEnableFIPS() bool
	GetIsoConfiguration() IsoConfiguration
	GetRawConfiguration() RawConfiguration
}

type Parser interface {
	ParseDefinition(data []byte) (Definition, error)
}

type Arch string

func (a Arch) Short() string {
	switch a {
	case ArchTypeX86:
		return "amd64"
	case ArchTypeARM:
		return "arm64"
	default:
		message := fmt.Sprintf("unknown arch: %s", a)
		panic(message)
	}
}

type Image struct {
	ImageType       string `yaml:"imageType"`
	Arch            Arch   `yaml:"arch"`
	BaseImage       string `yaml:"baseImage"`
	OutputImageName string `yaml:"outputImageName"`
}

type IsoConfiguration struct {
	InstallDevice string `yaml:"installDevice"`
}

type DiskSize string

func (d DiskSize) IsValid() bool {
	return diskSizeRegexp.MatchString(string(d))
}

func (d DiskSize) ToMB() int64 {
	if d == "" {
		return 0
	}

	s := diskSizeRegexp.FindStringSubmatch(string(d))
	if len(s) != 3 {
		panic("unknown disk size format")
	}

	quantity, err := strconv.Atoi(s[1])
	if err != nil {
		panic(fmt.Sprintf("invalid disk size: %s", string(d)))
	}

	sizeType := s[2]

	switch sizeType {
	case "M":
		return int64(quantity)
	case "G":
		return int64(quantity) * 1024
	case "T":
		return int64(quantity) * 1024 * 1024
	default:
		panic("unknown disk size type")
	}
}

type RawConfiguration struct {
	DiskSize                 DiskSize `yaml:"diskSize"`
	LUKSKey                  string   `yaml:"luksKey"`
	ExpandEncryptedPartition bool     `yaml:"expandEncryptedPartition"`
}

type Packages struct {
	NoGPGCheck      bool      `yaml:"noGPGCheck"`
	EnableExtras    bool      `yaml:"enableExtras"`
	PKGList         []string  `yaml:"packageList"`
	AdditionalRepos []AddRepo `yaml:"additionalRepos"`
	RegCode         string    `yaml:"sccRegistrationCode"`
}

type AddRepo struct {
	URL      string `yaml:"url"`
	Unsigned bool   `yaml:"unsigned"`
}

type OperatingSystemUser struct {
	Username          string   `yaml:"username"`
	UID               int      `yaml:"uid"`
	EncryptedPassword string   `yaml:"encryptedPassword"`
	SSHKeys           []string `yaml:"sshKeys"`
	PrimaryGroup      string   `yaml:"primaryGroup"`
	SecondaryGroups   []string `yaml:"secondaryGroups"`
	CreateHomeDir     bool     `yaml:"createHomeDir"`
}

type OperatingSystemGroup struct {
	Name string `yaml:"name"`
	GID  int    `yaml:"gid"`
}

type Systemd struct {
	Enable  []string `yaml:"enable"`
	Disable []string `yaml:"disable"`
}

type Suma struct {
	Host          string `yaml:"host"`
	ActivationKey string `yaml:"activationKey"`
}

type Time struct {
	Timezone         string           `yaml:"timezone"`
	NtpConfiguration NtpConfiguration `yaml:"ntp"`
}

type NtpConfiguration struct {
	ForceWait bool     `yaml:"forceWait"`
	Pools     []string `yaml:"pools"`
	Servers   []string `yaml:"servers"`
}

type Proxy struct {
	HTTPProxy  string   `yaml:"httpProxy"`
	HTTPSProxy string   `yaml:"httpsProxy"`
	NoProxy    []string `yaml:"noProxy"`
}

type EmbeddedArtifactRegistry struct {
	ContainerImages []ContainerImage `yaml:"images"`
	Registries      []Registry       `yaml:"registries"`
}

type ContainerImage struct {
	Name string `yaml:"name"`
}

type Registry struct {
	URI            string                 `yaml:"uri"`
	Authentication RegistryAuthentication `yaml:"authentication"`
}

type RegistryAuthentication struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Kubernetes struct {
	Version   string    `yaml:"version"`
	Network   Network   `yaml:"network"`
	Nodes     []Node    `yaml:"nodes"`
	Manifests Manifests `yaml:"manifests"`
	Helm      Helm      `yaml:"helm"`
}

type Network struct {
	APIHost string `yaml:"apiHost"`
	APIVIP4 string `yaml:"apiVIP"`
	APIVIP6 string `yaml:"apiVIP6"`
}

type Node struct {
	Hostname    string `yaml:"hostname"`
	Type        string `yaml:"type"`
	Initialiser bool   `yaml:"initializer"`
}

type Manifests struct {
	URLs []string `yaml:"urls"`
}

type Helm struct {
	Charts       []HelmChart      `yaml:"charts"`
	Repositories []HelmRepository `yaml:"repositories"`
}

type HelmChart struct {
	Name                  string   `yaml:"name"`
	ReleaseName           string   `yaml:"releaseName"`
	RepositoryName        string   `yaml:"repositoryName"`
	Version               string   `yaml:"version"`
	TargetNamespace       string   `yaml:"targetNamespace"`
	CreateNamespace       bool     `yaml:"createNamespace"`
	InstallationNamespace string   `yaml:"installationNamespace"`
	ValuesFile            string   `yaml:"valuesFile"`
	APIVersions           []string `yaml:"apiVersions"`
}

type HelmRepository struct {
	Name           string             `yaml:"name"`
	URL            string             `yaml:"url"`
	Authentication HelmAuthentication `yaml:"authentication"`
	PlainHTTP      bool               `yaml:"plainHTTP"`
	SkipTLSVerify  bool               `yaml:"skipTLSVerify"`
	CAFile         string             `yaml:"caFile"`
}

type HelmAuthentication struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
