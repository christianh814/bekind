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
	"os"

	"github.com/christianh814/bekind/pkg/helm"
	"github.com/christianh814/bekind/pkg/kind"
	"github.com/christianh814/bekind/pkg/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// convertMapInterface recursively converts map[interface{}]interface{} to map[string]interface{}
func convertMapInterface(data interface{}) interface{} {
	switch v := data.(type) {
	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for key, value := range v {
			strKey := fmt.Sprintf("%v", key)
			result[strKey] = convertMapInterface(value)
		}
		return result
	case []interface{}:
		for i, item := range v {
			v[i] = convertMapInterface(item)
		}
		return v
	default:
		return data
	}
}

// pullImages set to true by default
var pullImages bool = true

// HelmValues is the values provied in the configfile
type HelmValues struct {
	Name  string
	Value string
}

// HC is the extra helmcharts to install, if provided
var HC []struct {
	Url          string
	Repo         string
	Chart        string
	Release      string
	Namespace    string
	ValuesObject map[string]interface{}
	Wait         bool
	Version      string
}

// Set Default domain
var Domain string = "127.0.0.1.nip.io"

// Set the default Kind Image version
var KindImageVersion string

// ResetGlobalVars resets all global variables to their default state
// This is needed when running multiple profiles in sequence
func ResetGlobalVars() {
	log.Debug("Resetting global variables for next profile iteration")
	HC = nil // Clear the helm charts slice
	pullImages = true
	Domain = "127.0.0.1.nip.io"
	KindImageVersion = ""
}

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

		// check to see if the user wants to pull images before loading them into the cluster
		if viper.IsSet("loadDockerImages.pullImages") {
			pullImages = viper.GetBool("loadDockerImages.pullImages")
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
			log.Info("Using default KIND node image")

		}

		// Get images to load from the config file. NOTE: Images must exist on the host FIRST.
		dockerImages := viper.GetStringSlice("loadDockerImages.images")

		// Get post install manifests. NOTE: these need to be in YAML format currently
		// TODO: support for JSON formatted K8S Manifests
		postInstallManifests := viper.GetStringSlice("postInstallManifests")

		// Set the kindConfig as the config file for Viper
		kindConfig := viper.GetString("kindConfig")
		if len(kindConfig) == 0 {
			log.Fatal("Could not find kindConfig")
		}
		if err := viper.ReadConfig(bytes.NewBuffer([]byte(kindConfig))); err != nil {
			log.Fatal(err)
		}

		// Check to see if workers are being used. This is used to label the workers as such. This is based on inspecting the kindConfig
		var usesWorkers bool = false
		if len(viper.GetStringSlice("nodes")) > 1 {
			usesWorkers = true
		}

		// Check to see if the cluster name is set in the config file
		if viper.GetString("name") != "" {
			clusterName = viper.GetString("name")
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

		// Load images into the cluster
		if len(dockerImages) != 0 {
			log.Info("Loading Images in KIND cluster")
			if err := kind.LoadDockerImage(dockerImages, clusterName, pullImages); err != nil {
				log.Fatal(err)
			}
		}

		// Grab HelmCharts provided in the config file
		// Read YAML file directly to preserve key case sensitivity
		configFileToRead := cfgFile
		if configFileToRead == "" {
			// If no config file was specified via flag, check if viper loaded one
			configFileToRead = viper.ConfigFileUsed()
		}

		if configFileToRead != "" && viper.IsSet("helmCharts") {
			yamlData, err := os.ReadFile(configFileToRead)
			if err != nil {
				log.Fatal(err)
			}

			// Parse just the helmCharts section to preserve case
			var config struct {
				HelmCharts []struct {
					Url          string                 `yaml:"url"`
					Repo         string                 `yaml:"repo"`
					Chart        string                 `yaml:"chart"`
					Release      string                 `yaml:"release"`
					Namespace    string                 `yaml:"namespace"`
					ValuesObject map[string]interface{} `yaml:"valuesObject"`
					Wait         bool                   `yaml:"wait"`
					Version      string                 `yaml:"version"`
				} `yaml:"helmCharts"`
			}

			err = yaml.Unmarshal(yamlData, &config)
			if err != nil {
				log.Fatal(err)
			}

			// Convert to our HC format
			for _, chart := range config.HelmCharts {
				// Convert valuesObject from map[interface{}]interface{} to map[string]interface{}
				convertedValues := make(map[string]interface{})
				for k, v := range chart.ValuesObject {
					convertedValues[k] = convertMapInterface(v)
				}

				HC = append(HC, struct {
					Url          string
					Repo         string
					Chart        string
					Release      string
					Namespace    string
					ValuesObject map[string]interface{}
					Wait         bool
					Version      string
				}{
					Url:          chart.Url,
					Repo:         chart.Repo,
					Chart:        chart.Chart,
					Release:      chart.Release,
					Namespace:    chart.Namespace,
					ValuesObject: convertedValues,
					Wait:         chart.Wait,
					Version:      chart.Version,
				})
			}
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
				log.Infof("Installing Helm Chart %s/%s from %s", v.Repo, v.Chart, v.Url)

				if err := helm.Install(v.Namespace, v.Url, v.Repo, v.Chart, v.Release, v.Version, v.Wait, v.ValuesObject); err != nil {
					log.Fatal(err)
				}

				// Special conditions apply for Argo CD
				if v.Chart == "argo-cd" {

					// Get argo password
					argoSecret, err = client.CoreV1().Secrets("argocd").Get(context.TODO(), "argocd-initial-admin-secret", metav1.GetOptions{})
					if err != nil {
						if k8serrors.IsNotFound(err) {
							argoSecret.Data = map[string][]byte{
								"password": []byte("~* provided in helm chart *~"),
							}
						} else {
							log.Fatal(err)
						}
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

		// Set up a restconfig
		rc, err := utils.GetRestConfig("")
		if err != nil {
			log.Fatal(err)
		}

		// Load manifests into the cluster (if any)
		if len(postInstallManifests) != 0 {
			log.Info("Post Deployment Manifests")
			if err := utils.PostInstallManifests(postInstallManifests, context.TODO(), rc); err != nil {
				log.Warn("Issue with Post Install Manifests: ", err)
			}
		}

		// Save the bekind config to a secret
		log.Info("Saving bekind config to a secret in \"kube-public\"")
		err = utils.SaveBeKindConfig(rc, context.TODO(), "kube-public", "bekind-config")
		if err != nil {
			log.Fatal(err)
		}

		// Display Argo CD URL and password if it exists
		if argoUrl != "" {
			log.Infof("Argo CD is available at %s username: admin password: %s", argoUrl, argoPass)
		} else {
			log.Infof("KIND cluster %s is ready", clusterName)

		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
