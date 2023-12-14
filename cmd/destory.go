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
	"bytes"

	"github.com/christianh814/bekind/pkg/kind"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// destroyCmd represents the destory command
var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Aliases: []string{"delete", "del"},
	Short:   "Destroys the custom Kind cluster",
	Long: `Destroys a running custom Kind cluster. Currently
it only destroys the named cluster or it will destroy ones names "kind"
if one isn't named.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get clulster name from CLI
		clusterName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal(err)
		}

		// Set the kindConfig as the config file for Viper
		kindConfig := viper.GetString("kindConfig")
		if len(kindConfig) == 0 {
			log.Fatal("Could not find kindConfig")
		}
		viper.ReadConfig(bytes.NewBuffer([]byte(kindConfig)))

		// Check to see if the cluster name is set in the config file
		if viper.GetString("name") != "" {
			clusterName = viper.GetString("name")
		}

		// Set config file back to default for Viper
		viper.SetConfigFile(cfgFile)
		viper.ReadInConfig()

		log.Info("Destroying KIND cluster: ", clusterName)
		if err := kind.DeleteKindCluster(clusterName, ""); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
