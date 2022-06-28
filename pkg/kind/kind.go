package kind

import (
	"errors"

	"github.com/christianh814/bekind/pkg/utils"
	"sigs.k8s.io/kind/pkg/cluster"
)

var KindFullStack string = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: True
  podSubnet: "10.254.0.0/16"
  serviceSubnet: "172.30.0.0/16"
nodes:
- role: control-plane
- role: control-plane
- role: control-plane
- role: worker
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    listenAddress: 0.0.0.0
  - containerPort: 443
    hostPort: 443
    listenAddress: 0.0.0.0
- role: worker
- role: worker
`

var KindSingleNode string = `kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: True
  podSubnet: "10.254.0.0/16"
  serviceSubnet: "172.30.0.0/16"
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    listenAddress: 0.0.0.0
  - containerPort: 443
    hostPort: 443
    listenAddress: 0.0.0.0
`

// Set the default Kind Image version
var KindImageVersion string = "kindest/node:v1.24.2"

// We are using the same kind of provider for this whole package
var Provider *cluster.Provider = cluster.NewProvider(
	utils.GetDefaultRuntime(),
)

// CreateKindCluster creates KIND cluster
func CreateKindCluster(name string, installtype string) error {
	// Check to see what kind of install type we want
	switch installtype {
	case "":
		installtype = KindSingleNode
	case "full":
		installtype = KindFullStack
	case "single":
		installtype = KindSingleNode
	default:
		return errors.New("invalid install type")
	}

	// Create a KIND instance and write out the kubeconfig in the specified location
	err := Provider.Create(
		name,
		cluster.CreateWithRawConfig([]byte(installtype)),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithNodeImage(KindImageVersion),
	)

	if err != nil {
		return err
	}

	return nil
}
