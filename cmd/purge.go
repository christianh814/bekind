/*
Copyright © 2024 Christian Hernandez <christian@chernand.io>

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
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// purgeCmd represents the purge command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Deletes all running Kind clusters on the system",
	Long: `This command will delete all running Kind clusters on the system, regardless of the name and
and regardless if bekind created them or not. This command is destructive and will delete
all running KIND clusters.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get the value from the CLI
		confirm, err := cmd.Flags().GetBool("confirm")
		if err != nil {
			log.Fatal(err)
		}

		if confirm {
			// Delete all clusters if the user confirms
			purge()
		} else {
			// Ask if the user wants to delete everything var ans string
			var ans string
			fmt.Printf("Are you sure you want to delete all KIND clusters on the system? [y/N]: ")
			fmt.Scan(&ans)

			if ans == "y" || ans == "Y" {
				purge()
			} else {
				log.Info("Exiting")
			}

		}
	},
}

func init() {
	rootCmd.AddCommand(purgeCmd)

	purgeCmd.PersistentFlags().BoolP("confirm", "c", false, "Confirm deleting all Kind clusters on the system")
}

func purge() {
	log.Info("Deleting all KIND clusters on the system")
	err := kind.DeleteAllKindClusters("")
	if err != nil {
		log.Fatal(err)
	}
}
