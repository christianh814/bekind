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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Set ProfileDir
// TODO: use os.UserHomeDir()
var ProfileDir = os.Getenv("HOME") + "/.bekind/profiles"

// runCmd runs a profile
var runCmd = &cobra.Command{
	Use:               "run <profile>",
	Args:              cobra.MatchAll(cobra.MinimumNArgs(1)),
	Short:             "Runs the specified profile",
	Long:              profileLongHelp(),
	ValidArgsFunction: profileValidArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Set Config file based on the profile
		viper.SetConfigFile(ProfileDir + "/" + args[0] + "/config.yaml")

		// Read the config file, Only displaying an error if there was a problem reading the file.
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				log.Fatal(err)
			}
		}

		// If the view flag is set, show the config
		if view, _ := cmd.Flags().GetBool("view"); view {
			showconfigCmd.Run(cmd, []string{})
			os.Exit(0)
		}

		// Run the profile
		startCmd.Run(cmd, []string{})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add a view flag that takes a string argument
	runCmd.Flags().BoolP("view", "v", false, "View the profile configuration")
}

// profileValidArgs returns a list of profiles for tab completion
func profileValidArgs(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	p, _ := getProfileNames()
	return p, cobra.ShellCompDirectiveNoFileComp
}

// getProfileNames returns a list of profile names based on the files in the profile directory
func getProfileNames() ([]string, error) {
	p := []string{}
	e, err := os.ReadDir(ProfileDir)

	if err != nil {
		return p, err
	}

	for _, entry := range e {
		p = append(p, entry.Name())
	}

	return p, nil
}

// profileLongHelp returns the long help for the profile command
func profileLongHelp() string {
	return `You can use "run" to run the specified profile. Profiles needs to be
stored in the ~/.bekind/profiles/{{name}} directory.

The profile directory should contain a config.yaml file. For example:

~/.bekind/profiles/{{name}}/config.yaml`
}
