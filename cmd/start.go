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

	"github.com/christianh814/bekind/pkg/helm"
	"github.com/christianh814/bekind/pkg/kind"
	"github.com/christianh814/bekind/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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
	Wait      bool
	Version   string
}

// Set Default domain
var Domain string = "127.0.0.1.nip.io"

// Set the default Kind Image version
var KindImageVersion string = "kindest/node:v1.28.0"

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts a custom Kind cluster",
	Long: `This command starts a custom Kind cluster based 
on the configuration file that is passed`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting KIND cluster")

		// Get clulster name from CLI
		clusterName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatal(err)
		}

		// Get "domain" from the config file if it exists using viper
		// Leaving this here although not using "domain" anymore, it might
		// be useful in the future.
		if viper.GetString("domain") != "" {
			Domain = viper.GetString("domain")
			log.Warn("Using custom domain")
		}

		// Get "kindImageVersion" from the config file if it exists using viper
		if viper.GetString("kindImageVersion") != "" {
			KindImageVersion = viper.GetString("kindImageVersion")
			log.Warn("Using custom KIND node image " + KindImageVersion)
		} else {
			log.Info("Using KIND node image " + KindImageVersion)

		}

		// Get images to load from the config file. NOTE: Images must exist on the host FIRST.
		dockerImages := viper.GetStringSlice("loadDockerImages")

		// Set the kindConfig as the config file for Viper
		kindConfig := viper.GetString("kindConfig")
		if len(kindConfig) == 0 {
			log.Fatal("Could not find kindConfig")
		}
		viper.ReadConfig(bytes.NewBuffer([]byte(kindConfig)))

		// Check to see if workers are being used. This is used to label the workers as such. This is based on inspecting the kindConfig
		var usesWorkers bool = false
		if len(viper.GetStringSlice("nodes")) > 1 {
			usesWorkers = true
		}

		// Set config file back to default for Viper
		viper.SetConfigFile(cfgFile)
		viper.ReadInConfig()

		// Try and start the kind cluster
		err = kind.CreateKindCluster(clusterName, KindImageVersion)
		if err != nil {
			log.Fatal(err)
		}

		// Get the client from the new Kubernetes clusters
		client, err := utils.NewClient("")
		if err != nil {
			log.Fatal(err)
		}

		// If not a single node then label the workers as such
		if usesWorkers {
			log.Info("Labeling workers")
			err = utils.LabelWorkers(client)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Grab HelmCharts provided in the config file
		err = viper.UnmarshalKey("helmCharts", &HC)
		if err != nil {
			log.Fatal(err)
		}

		// Special conditions for Argo CD
		var argoSecret *v1.Secret
		var argoIngress *networkingv1.Ingress
		var argoUrl string
		var argoPass string

		// Install Helm Charts if any exist in the config file
		if len(HC) != 0 {
			// Range over the helmCharts and try to install them
			// 	TODO: Currently it's garbage in garbage out, if the user provides a bad chart it will fail
			for _, v := range HC {
				// Install HelmChart
				HelmArgs := map[string]string{
					// comma seperated values to set
					"set": fmt.Sprintf(v.Args),
				}
				log.Infof("Installing Helm Chart %s/%s from %s", v.Repo, v.Chart, v.Url)

				if err := helm.Install(v.Namespace, v.Url, v.Repo, v.Chart, v.Release, v.Version, v.Wait, HelmArgs); err != nil {
					log.Fatal(err)
				}

				// Special conditions apply for Argo CD
				if v.Chart == "argo-cd" {

					// Get argo password
					argoSecret, err = client.CoreV1().Secrets("argocd").Get(context.TODO(), "argocd-initial-admin-secret", metav1.GetOptions{})
					if err != nil {
						log.Fatal(err)
					}

					// Get argo ingress
					argoIngress, err = client.NetworkingV1().Ingresses("argocd").Get(context.TODO(), "argocd-server", metav1.GetOptions{})
					if err != nil {
						log.Fatal(err)
					}

					// Save information for later use
					argoUrl = fmt.Sprintf("https://%s", argoIngress.Spec.Rules[0].Host)
					argoPass = string(argoSecret.Data["password"])

				}

			}
		}

		// Load images into the cluster
		if len(dockerImages) != 0 {
			log.Info("Loading Images in KIND cluster")
			if err := kind.LoadDockerImage(dockerImages, clusterName); err != nil {
				log.Fatal(err)
			}
		}

		// Display Argo CD URL and password if it exists
		if argoUrl != "" {
			log.Infof("Argo CD is available at %s username: admin password %s", argoUrl, argoPass)
		} else {
			log.Infof("KIND cluster %s is ready", clusterName)

		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
