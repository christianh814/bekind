/*
Copyright © 2022 Christian Hernandez christian@chernand.io

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/usrbinkat/bekinder/pkg/clusterops"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a Kubernetes cluster using KIND",
	Long: `Create a Kubernetes cluster using KIND (Kubernetes in Docker).
This command supports creating different types of clusters, such as single-node, multi-node, and custom.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return createCluster()
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	// Flags for the create command
	createCmd.Flags().StringP("name", "n", "bekind", "Name of the KIND cluster")
	createCmd.Flags().StringP("installtype", "t", "", "Cluster type: 'single', 'full', or 'custom'")
	createCmd.Flags().StringP("kindimage", "i", "", "Specify a custom KIND node image")

	// Bind flags to Viper configuration
	viper.BindPFlag("clusterName", createCmd.Flags().Lookup("name"))
	viper.BindPFlag("installType", createCmd.Flags().Lookup("installtype"))
	viper.BindPFlag("kindImage", createCmd.Flags().Lookup("kindimage"))
}

// createCluster creates a new KIND cluster based on the given parameters
func createCluster() error {
	clusterName := viper.GetString("clusterName")
	installType := viper.GetString("installType")
	kindImage := viper.GetString("kindImage")

	err := clusterops.CreateKindCluster(clusterName, installType, kindImage)
	if err != nil {
		return fmt.Errorf("failed to create KIND cluster: %w", err)
	}

	fmt.Printf("KIND cluster '%s' created successfully\n", clusterName)
	return nil
}
