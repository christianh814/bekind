package cluster

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	clusterops "github.com/usrbinkat/bekind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster"
)

var Provider *cluster.Provider = cluster.NewProvider()

// CreateKindCluster creates a KIND cluster using the provided configuration
func CreateKindCluster(name string, installType string, kindImage string) error {
	config, err := clusterops.GetKindConfig(installType, kindImage)
	if err != nil {
		return errors.Wrap(err, "failed to get KIND configuration")
	}

	err = Provider.Create(name, config)
	if err != nil {
		return errors.Wrap(err, "failed to create KIND cluster")
	}

	// Save the kubeconfig for the newly created cluster
	kubeconfig, err := Provider.KubeConfig(name, false)
	if err != nil {
		return errors.Wrap(err, "failed to get kubeconfig for the KIND cluster")
	}

	err = clusterops.SaveKubeConfig(kubeconfig, name)
	if err != nil {
		return errors.Wrap(err, "failed to save kubeconfig")
	}

	// Patch the cluster with custom configurations
	err = clusterops.PatchCluster(viper.GetString("kubeconfig"))
	if err != nil {
		return errors.Wrap(err, "failed to patch the cluster")
	}

	return nil
}

// DeleteKindCluster deletes a KIND cluster with the specified name
func DeleteKindCluster(name string) error {
	err := Provider.Delete(name, "")
	if err != nil {
		return errors.Wrap(err, "failed to delete the KIND cluster")
	}

	return nil
}
