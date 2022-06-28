package utils

import (
	"os"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

// GetDefault selected the default runtime from the environment override
func GetDefaultRuntime() cluster.ProviderOption {
	switch p := os.Getenv("KIND_EXPERIMENTAL_PROVIDER"); p {
	case "":
		return nil
	case "podman":
		log.Warn("using podman due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithPodman()
	case "docker":
		log.Warn("using docker due to KIND_EXPERIMENTAL_PROVIDER")
		return cluster.ProviderWithDocker()
	default:
		log.Warnf("ignoring unknown value %q for KIND_EXPERIMENTAL_PROVIDER", p)
		return nil
	}
}

// SetKubeConfig sets the kubeconfig path
func SetKubeConfig(kubeconfig string) (string, error) {
	// Set up KubeConfig Globally
	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}
	if kubeconfig == "" {
		return clientcmd.RecommendedHomeFile, nil // use default path(.kube/config)
	}

	return kubeconfig, nil
}
