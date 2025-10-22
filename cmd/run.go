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
	"os"
	"path/filepath"

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
		// Set ProfileDir based on flag if provided by user
		profileDirFlag, err := cmd.Flags().GetString("profile-dir")
		if err == nil && profileDirFlag != "" {
			ProfileDir = profileDirFlag
		}

		// Look for all yaml files in the profile directory
		configFiles, err := filepath.Glob(filepath.Join(ProfileDir+"/"+args[0], "*.yaml"))
		if err != nil {
			log.Fatal(err)
		}

		// If no config files are found, exit with an error
		if len(configFiles) == 0 {
			log.Fatalf("No config files found in profile directory: %s", ProfileDir+"/"+args[0])
		}

		// Get view flag
		view, err := cmd.Flags().GetBool("view")
		if err != nil {
			log.Fatal(err)
		}

		// Iterate over all config files and run the profile for each one
		for _, configFile := range configFiles {
			// Set Config file based on the profile
			viper.SetConfigFile(configFile)

			// Read the config file, Only displaying an error if there was a problem reading the file.
			if err := viper.ReadInConfig(); err != nil {
				if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
					log.Fatal(err)
				}
			}

			// If the view flag is set, show the config
			if view {
				// Crude, but it works
				// TODO: Create a "BeKindStack" YAML struct and use that to display the config
				fmt.Println("---")
				showconfigCmd.Run(cmd, []string{})
			} else {
				// I assume you want to "Run the profile"
				if !view {
					log.Info("Running profile: ", args[0], " with config file: ", filepath.Base(configFile))
					startCmd.Run(cmd, []string{})
				}

			}

			// Clear viper config for next iteration
			viper.Reset()

			// Reset global variables from start.go to prevent state leakage between iterations
			ResetGlobalVars()
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add a view flag that takes a string argument
	runCmd.Flags().BoolP("view", "v", false, "View the profile configuration")

	// Add a profile-dir flag that takes a string argument use StringVar
	runCmd.Flags().StringVarP(&ProfileDir, "profile-dir", "p", ProfileDir, "Directory where profiles are stored")

	// Mark profile-dir flag and config flag as mutually exclusive
	runCmd.MarkFlagsMutuallyExclusive("profile-dir", "config")
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

The profile directory should contain a config file in YAML format. For example:

~/.bekind/profiles/{{name}}/config.yaml

NOTE: You can have multiple YAML configurations in the same profile directory.

If you're specifying a directory, you must use base name of the directory.

For example, if your config file is stored under /tmp/foo then you'd run:

	  bekind run foo --profile-dir /tmp
	  
You can also use the --view flag to view the configuration of the profile without running it.`
}
