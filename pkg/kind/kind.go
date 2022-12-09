package kind

import (
	"errors"

	"github.com/christianh814/bekind/pkg/utils"
	"github.com/spf13/viper"
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
  kubeadmConfigPatches:
  - |
    kind: JoinConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "haproxy=ingresshost"
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
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "nginx=ingresshost"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    listenAddress: 0.0.0.0
  - containerPort: 443
    hostPort: 443
    listenAddress: 0.0.0.0
`

// Set the default Kind Image version
var KindImageVersion string = "kindest/node:v1.26.0"

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
	case "custom":
		installtype = viper.GetString("kindConfig")
	default:
		return errors.New("invalid install type")
	}

	// If a config file is given, try to use that. Garbage in, garbage out though
	suppliedConfig := viper.GetString("kindConfig")
	if suppliedConfig != "" {
		installtype = suppliedConfig
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

// DeleteKindCluster deletes KIND cluster based on the name given
func DeleteKindCluster(name string, cfg string) error {
	err := Provider.Delete(name, cfg)

	if err != nil {
		return err
	}

	return nil

}
