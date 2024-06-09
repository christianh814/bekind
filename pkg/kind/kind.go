package kind

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/christianh814/bekind/pkg/utils"
	"github.com/spf13/viper"
	kindConfig "sigs.k8s.io/kind/pkg/apis/config/defaults"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/nodes"
	"sigs.k8s.io/kind/pkg/cluster/nodeutils"
	"sigs.k8s.io/kind/pkg/fs"
)

type (
	imageTagFetcher func(nodes.Node, string) (map[string]bool, error)
)

// We are using the same kind of provider for this whole package
var Provider *cluster.Provider = cluster.NewProvider(
	utils.GetDefaultRuntime(),
)

// CreateKindCluster creates KIND cluster
func CreateKindCluster(name string, kindImage string) error {
	// If a config file is given, try to use that. Garbage in, garbage out though
	// check if the config is actually there
	installConfig := viper.GetString("kindConfig")
	if installConfig == "" {
		return errors.New("no valid config found")
	}

	// If the image is not given, use the default image
	if kindImage == "" {
		kindImage = kindConfig.Image
	}

	// Create a KIND instance and write out the kubeconfig in the specified location
	err := Provider.Create(
		name,
		cluster.CreateWithRawConfig([]byte(installConfig)),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithNodeImage(kindImage),
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

// DeleteAllKindClusters deletes all KIND clusters
func DeleteAllKindClusters(cfg string) error {
	// List all clusters
	clusters, err := ListKindClusters()
	if err != nil {
		return err
	}

	// Loop through all clusters and delete them
	for _, cluster := range clusters {
		err := DeleteKindCluster(cluster, cfg)
		if err != nil {
			return err
		}
	}

	// If we are here we should be okay
	return nil
}

// ListKindClusters lists KIND clusters
func ListKindClusters() ([]string, error) {
	return Provider.List()
}

// LoadDockerImage loads a docker image into the KIND cluster
func LoadDockerImage(images []string, clustername string, pull bool) error {
	// If we are pulling the images, do that first
	if pull {
		err := pullImages(images)
		if err != nil {
			return err
		}
	}
	// Get the list of nodes in the cluster
	nodes, err := Provider.ListNodes(clustername)
	if err != nil {
		return err
	}

	// If no nodes were returned, we have a problem
	if len(nodes) == 0 {
		return errors.New("no nodes found")
	}

	// Setup the tar path where the images will be saved
	dir, err := fs.TempDir("", "images-tar")
	if err != nil {
		return errors.New("failed to create tempdir")
	}

	defer os.RemoveAll(dir)
	imagesTarPath := filepath.Join(dir, "images.tar")
	// Save the images into a tar
	err = save(images, imagesTarPath)
	if err != nil {
		return err
	}

	// Load the images on the selected nodes
	for _, selectedNode := range nodes {
		selectedNode := selectedNode // capture loop variable
		return loadImage(imagesTarPath, selectedNode)
	}

	// If we are here we should be okay
	return nil
}

// save saves images to dest, as in `docker save`
func save(images []string, dest string) error {
	commandArgs := append([]string{"save", "-o", dest}, images...)
	return exec.Command("docker", commandArgs...).Run()
}

// loadImage loads an image tarball onto a node
func loadImage(imageTarName string, node nodes.Node) error {
	f, err := os.Open(imageTarName)
	if err != nil {
		return errors.New("failed to open image")
	}
	defer f.Close()
	return nodeutils.LoadImageArchive(node, f)
}

// pullImages pulls images locally so that they can be loaded into the cluster
func pullImages(images []string) error {
	for _, image := range images {
		err := exec.Command("docker", "pull", image).Run()
		if err != nil {
			return errors.New("failed to pull image: " + image)
		}
	}
	return nil
}
