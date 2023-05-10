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
	"context"
	"fmt"
	"time"

	"github.com/christianh814/bekind/pkg/helm"
	"github.com/christianh814/bekind/pkg/kind"
	"github.com/christianh814/bekind/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HC is the extra helmcharts to install, if provided
var HC []struct {
	Url       string
	Repo      string
	Chart     string
	Release   string
	Namespace string
	Args      string
}

// disablecni disables the CNI plugin in the kind cluster
var disablecni bool = false

// Set Default domain
var Domain string = "127.0.0.1.nip.io"

// Set the default Kind Image version
var KindImageVersion string = "kindest/node:v1.26.3"

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a custom Kind cluster",
	Long: `This command starts a custom Kind cluster. Currently
it installs Argo CD and an HAProxy Ingress controller.`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting KIND cluster")

		// Get clulster name from CLI
		clusterName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal(err)
		}

		// Get clulster type from CLI
		var clusterType string
		isSingleNode, err := cmd.Flags().GetBool("single")
		if err != nil {
			log.Fatal(err)
		}

		// Get "domain" from the config file if it exists using viper
		if viper.GetString("domain") != "" {
			Domain = viper.GetString("domain")
			log.Warn("Using custom domain for ingress")
		}

		// Get "kindImageVersion" from the config file if it exists using viper
		if viper.GetString("kindImageVersion") != "" {
			KindImageVersion = viper.GetString("kindImageVersion")
			log.Warn("Using custom KIND node image ")
		}

		// Set Cluster type
		if isSingleNode {
			clusterType = "single"
		} else {
			clusterType = "full"
		}

		// Get the custom kind config from the config file
		kindConfig := viper.GetString("kindConfig")
		if kindConfig != "" {
			clusterType = "custom"
			log.Warn("Using custom kind config")
		}

		// Set the kindConfig as the config file for Viper
		viper.ReadConfig(bytes.NewBuffer([]byte(kindConfig)))

		// Check to see if the CNI is disabled
		if viper.Get("networking.disableDefaultCNI") != nil {
			disablecni = viper.Get("networking.disableDefaultCNI").(bool)
		}

		// Set config file back to default for Viper
		viper.SetConfigFile(cfgFile)
		viper.ReadInConfig()

		// Do we install argocd? Get from CLI
		installArgo, err := cmd.Flags().GetBool("argocd")
		if err != nil {
			log.Fatal(err)
		}

		// Try and start the kind cluster
		err = kind.CreateKindCluster(clusterName, clusterType, KindImageVersion)
		if err != nil {
			log.Fatal(err)
		}

		// Get the client from the new Kubernetes clusters
		client, err := utils.NewClient("")
		if err != nil {
			log.Fatal(err)
		}

		// If not a single node then label the workers as such
		if !isSingleNode {
			log.Info("Labeling workers")
			err = utils.LabelWorkers(client)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Grab any extra HelmCharts provided in the config file
		viper.UnmarshalKey("helmCharts", &HC)

		// Install Default bekind CNI if CNI is disabled
		if disablecni {

			// Install Calico CNI
			var (
				calicoUrl         = "https://projectcalico.docs.tigera.io/charts"
				calicoRepoName    = "projectcalico"
				calicoReleaseName = "calico"
				calicoChartName   = "tigera-operator"
				calicoNamespace   = "calico-system"
				calicoHelmArgs    = map[string]string{
					"set": `installation.calicoNetwork.ipPools[0].blockSize=26,installation.calicoNetwork.ipPools[0].cidr=10.254.0.0/16,installation.calicoNetwork.ipPools[0].encapsulation=VXLANCrossSubnet,installation.calicoNetwork.ipPools[0].natOutgoing=Enabled,installation.calicoNetwork.ipPools[0].nodeSelector=all()`,
				}
			)
			log.Info("Installing Calico CNI")
			if err := helm.Install(calicoNamespace, calicoUrl, calicoRepoName, calicoChartName, calicoReleaseName, calicoHelmArgs); err != nil {
				log.Fatal(err)
			}

			// Wait for Calico rollout to happen
			log.Info("Waiting for Calico rollout")
			if err = utils.WaitForDeployment(client, calicoNamespace, "calico-typha", 600*time.Second); err != nil {
				log.Fatal(err)
			}

		} else {
			log.Info("Installing Default KIND CNI")
		}

		// Install ingress controller
		var (
			ingressURL         = "https://kubernetes.github.io/ingress-nginx"
			ingressRepoName    = "ingress-nginx"
			ingressChartName   = "ingress-nginx"
			ingressReleaseName = "nginx-ingress"
			ingressNamespace   = "ingress-controller"
			ingressHelmArgs    = map[string]string{
				// comma seperated values to set
				"set": `controller.hostNetwork=true,controller.nodeSelector.nginx=ingresshost,controller.service.type=ClusterIP,controller.service.externalTrafficPolicy=,controller.extraArgs.enable-ssl-passthrough=`,
			}
		)
		log.Info("Installing ingress controller")
		if err := helm.Install(ingressNamespace, ingressURL, ingressRepoName, ingressChartName, ingressReleaseName, ingressHelmArgs); err != nil {
			log.Fatal(err)
		}

		// Wait for Ingress Controller rollout to happen
		log.Info("Waiting for Ingress rollout")
		if err = utils.WaitForDeployment(client, ingressNamespace, "nginx-ingress-ingress-nginx-controller", 600*time.Second); err != nil {
			log.Fatal(err)
		}

		// Install Argo CD
		if installArgo {

			// Install ingress controller
			var (
				argoURL         = "https://argoproj.github.io/argo-helm"
				argoRepoName    = "argo"
				argoChartName   = "argo-cd"
				argoReleaseName = "argocd"
				argoNamespace   = "argocd"
				argoHelmArgs    = map[string]string{
					// comma seperated values to set
					"set": `server.ingress.enabled=true,server.ingress.hosts[0]=argocd.` + Domain + `,server.ingress.ingressClassName="nginx",server.ingress.https=true,server.ingress.annotations."nginx\.ingress\.kubernetes\.io/ssl-passthrough"=true,server.ingress.annotations."nginx\.ingress\.kubernetes\.io/force-ssl-redirect"=true`,
				}
			)
			log.Info("Installing Argo CD")
			if err := helm.Install(argoNamespace, argoURL, argoRepoName, argoChartName, argoReleaseName, argoHelmArgs); err != nil {
				log.Fatal(err)
			}

			// Wait for Argo CD rollout to happen
			log.Info("Waiting for Argo CD rollout")
			if err = utils.WaitForDeployment(client, argoNamespace, "argocd-server", 600*time.Second); err != nil {
				log.Fatal(err)
			}

		} else {
			log.Info("Skipping Argo CD installation")
		}

		// Get argo password
		argoSecret, err := client.CoreV1().Secrets("argocd").Get(context.TODO(), "argocd-initial-admin-secret", metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}

		// Get argo ingress
		argoIngress, err := client.NetworkingV1().Ingresses("argocd").Get(context.TODO(), "argocd-server", metav1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}

		argoUrl := fmt.Sprintf("https://%s", argoIngress.Spec.Rules[0].Host)
		argoPass := string(argoSecret.Data["password"])

		// Install Helm Charts if any exist in the config file
		if len(HC) != 0 {
			log.Info("Installing Additional HelmCharts from config file")
			// Range over the helmCharts and try to install them
			// 	TODO: Currently it's garbage in garbage out, if the user provides a bad chart it will fail
			for _, v := range HC {

				// Install HelmChart
				HelmArgs := map[string]string{
					// comma seperated values to set
					"set": fmt.Sprintf(v.Args),
				}
				if err := helm.Install(v.Namespace, v.Url, v.Repo, v.Chart, v.Release, HelmArgs); err != nil {
					log.Fatal(err)
				}
			}
			//
		}

		//
		log.Infof("Argo CD is available at %s username: admin password %s", argoUrl, argoPass)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")
	startCmd.PersistentFlags().Bool("single", false, "Install a single instance of the kind cluster")
	startCmd.PersistentFlags().Bool("argocd", true, "Install Argo CD")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
