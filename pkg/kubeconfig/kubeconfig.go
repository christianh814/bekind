package kubeconfig

import (
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetKubeConfigPath returns the kubeconfig path.
func GetKubeConfigPath() string {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	if kubeConfigPath == "" {
		kubeConfigPath = clientcmd.RecommendedHomeFile // use default path (.kube/config)
	}
	return kubeConfigPath
}

// GetKubeConfig returns a *rest.Config for the given kubeconfig path.
func GetKubeConfig(kubeConfigPath string) (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}

// NewClient returns a kubernetes.Interface.
func NewClient(kubeConfigPath string) (kubernetes.Interface, error) {
	kubeConfig, err := GetKubeConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(kubeConfig)
}
