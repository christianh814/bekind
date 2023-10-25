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

	"github.com/christianh814/bekind/pkg/kind"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List running bekind instances",
	Long:    `List running bekind instances, it will also list instances that were created by KIND directly.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get a list of KIND clusters
		clusters, err := kind.ListKindClusters()
		if err != nil {
			log.Fatal(err)
		}

		// Check to see if there are any clusters
		if len(clusters) == 0 {
			log.Info("No clusters found")
		}

		// list clusters
		for _, cluster := range clusters {
			fmt.Println(cluster)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
