package cluster

import (
	"errors"

	"github.com/spf13/viper"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/internal/runtime"
)

var Provider *cluster.Provider = cluster.NewProvider(
	cluster.ProviderWithDefault(runtime.ProviderFor(runtime.Docker)),
)

func CreateKindCluster(name string, installtype string, kindImage string) error {
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

	suppliedConfig := viper.GetString("kindConfig")
	if suppliedConfig != "" {
		installtype = suppliedConfig
	}

	err := Provider.Create(
		name,
		cluster.CreateWithRawConfig([]byte(installtype)),
		cluster.CreateWithDisplayUsage(false),
		cluster.CreateWithDisplaySalutation(false),
		cluster.CreateWithNodeImage(kindImage),
	)

	if err != nil {
		return err
	}

	return nil
}

func DeleteKindCluster(name string, cfg string) error {
	err := Provider.Delete(name, cfg)

	if err != nil {
		return err
	}

	return nil
}
