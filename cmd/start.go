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
	"github.com/christianh814/bekind/pkg/helm"
	"github.com/christianh814/bekind/pkg/kind"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting KIND cluster")

		// Get clulster name from CLI
		clusterName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal(err)
		}
		// Get clulster type from CLI
		clusterType, err := cmd.Flags().GetString("type")
		if err != nil {
			log.Fatal(err)
		}

		// Try and start the kind cluster
		err = kind.CreateKindCluster(clusterName, clusterType)
		if err != nil {
			log.Fatal(err)
		}

		// Install Calico CNI
		var (
			calicoUrl         = "https://projectcalico.docs.tigera.io/charts"
			calicoRepoName    = "projectcalico"
			calicoReleaseName = "calico"
			calicoChartName   = "tigera-operator"
			calicoNamespace   = "calico-system"
			calicoHelmArgs    = map[string]string{
				"set": `installation.calicoNetwork.ipPools[0].blockSize=26,installation.calicoNetwork.ipPools[0].cidr="10.254.0.0/16",installation.calicoNetwork.ipPools[0].encapsulation="VXLANCrossSubnet",installation.calicoNetwork.ipPools[0].natOutgoing="Enabled",installation.calicoNetwork.ipPools[0].nodeSelector="all()"`,
			}
		)
		log.Info("Installing Calico CNI")
		if err := helm.Install(calicoNamespace, calicoUrl, calicoRepoName, calicoChartName, calicoReleaseName, calicoHelmArgs); err != nil {
			log.Fatal(err)
		}

		// Install ingress controller

		/*
			// Install ingress controller
			var (
				url         = "https://haproxy-ingress.github.io/charts"
				repoName    = "ingress"
				chartName   = "haproxy-ingress"
				releaseName = "ingress"
				namespace   = "ingress-controller"
				helmArgs    = map[string]string{
					// comma seperated values to set
					"set": "controller.hostNetwork=true,controller.nodeSelector.haproxy=ingresshost,controller.service.type=ClusterIP,controller.service.externalTrafficPolicy=",
				}
			)
			log.Info("Installing ingress controller")
			//TODO: create namespace
			if err := helm.Install(); err != nil {
				log.Fatal(err)
			}
		*/

		//
		log.Info("Install Complete")
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
	startCmd.PersistentFlags().StringP("name", "n", "kind", "The name of the kind instance")
	startCmd.PersistentFlags().StringP("type", "t", "full", "The type of install to use for the kind instance ('full' or 'single')")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
