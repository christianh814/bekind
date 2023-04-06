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

	"github.com/usrbinkat/bekind/pkg/kubeconfig"
)

// showConfigCmd represents the showconfig command
var showConfigCmd = &cobra.Command{
	Use:   "showconfig",
	Short: "Displays the current configuration",
	Long:  `Displays the current configuration used by bekind.`,
	Run: func(cmd *cobra.Command, args []string) {
		showConfig()
	},
}

func init() {
	rootCmd.AddCommand(showConfigCmd)
}

// showConfig prints the current configuration
func showConfig() {
	// Get all the keys from the Viper configuration
	keys := viper.AllKeys()

	fmt.Println("Current configuration:")
	for _, key := range keys {
		fmt.Printf("%s: %v\n", key, viper.Get(key))
	}

	// Display the configuration file used
	fmt.Printf("\nConfiguration file used: %s\n", kubeconfig.GetKubeConfigPath())
}
