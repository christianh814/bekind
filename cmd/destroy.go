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

// destroyCmd represents the destroy command
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a Kubernetes cluster created with KIND",
	Long:  `Destroy a Kubernetes cluster created with KIND. This command deletes the specified cluster and cleans up all related resources.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return destroyCluster()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)

	// Flags for the destroy command
	destroyCmd.Flags().StringP("name", "n", "bekind", "Name of the KIND cluster to destroy")

	// Bind flags to Viper configuration
	viper.BindPFlag("clusterName", destroyCmd.Flags().Lookup("name"))
}

// destroyCluster destroys a KIND cluster based on the given parameters
func destroyCluster() error {
	clusterName := viper.GetString("clusterName")

	err := clusterops.DeleteKindCluster(clusterName, "")
	if err != nil {
		return fmt.Errorf("failed to destroy KIND cluster: %w", err)
	}

	fmt.Printf("KIND cluster '%s' destroyed successfully\n", clusterName)
	return nil
}
