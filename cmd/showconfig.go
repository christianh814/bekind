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
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

// showconfigCmd represents the showconfig command
var showconfigCmd = &cobra.Command{
	Use:     "showconfig",
	Aliases: []string{"sc", "showConfig", "configShow"},
	Short:   "Prints out the config that will be used",
	Long: `Prints out the config that will be used by
bekind to set up your local Kind cluster.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Marshal in the entire config file int a byteslice and check for errors
		byteSlice, err := yaml.Marshal(viper.AllSettings())
		if err != nil {
			log.Fatal(err)
		}

		// Print it out
		fmt.Print(string(byteSlice))

	},
}

func init() {
	rootCmd.AddCommand(showconfigCmd)
}
