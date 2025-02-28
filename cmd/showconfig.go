/*
Copyright © 2023 Christian Hernandez <christian@chernand.io>

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
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/christianh814/bekind/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// showconfigCmd represents the showconfig command
var showconfigCmd = &cobra.Command{
	Use:     "showconfig",
	Aliases: []string{"sc", "showConfig", "configShow"},
	Short:   "Prints out the config that will be used",
	Long: `Prints out the config that will be used by bekind to set up your local Kind cluster.

In the case you use --system it will print out the config that is saved on the cluster. This is
usually the "config.yaml" value in the "bekind-config" secret in the "kube-public" namespace.`,
	Run: func(cmd *cobra.Command, args []string) {
		byteSlice := []byte{}
		err := error(nil)

		// Check to see if the user wants to print out the config saved on the cluster
		if viper.GetBool("system") {
			// Get the config from the cluster
			rc, _ := utils.GetRestConfig("")
			byteSlice, err = utils.GetBeKindConfig(rc, context.TODO(), "kube-public", "bekind-config")
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// Marshal in the entire config file int a byteslice and check for errors
			byteSlice, err = yaml.Marshal(viper.AllSettings())
			if err != nil {
				log.Fatal(err)
			}

		}

		// Print it out
		fmt.Print(string(byteSlice))

	},
}

func init() {
	rootCmd.AddCommand(showconfigCmd)
	showconfigCmd.Flags().BoolP("system", "s", false, "Prints out the config saved on the cluster")
}
