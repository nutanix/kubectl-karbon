/*
Package cmd root of the tools
Copyright © 2021 Christophe Jauffret <christophe@nutanix.com>

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

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var verbose bool
var debug bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "kubectl-karbon",
	Short: "Karbon Plugin for kubectl.",
	// Long:  `Karbon Plugin for kubectl.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// cobra.CheckErr(rootCmd.Execute())
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "karbon plugin config file (default ~/.kubectl-karbon.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "print verbose logging information")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "print debug logging information")
	rootCmd.PersistentFlags().Int("request-timeout", 30, "request timeout in seconds for HTTP client")

	userHomeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	defaultKubeconfig := fmt.Sprintf("%s/.kube/config", userHomeDir)

	rootCmd.PersistentFlags().String("kubeconfig", defaultKubeconfig, "path to the kubeconfig file to use for CLI requests")
	viper.BindPFlag("kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name ".kubectl-karbon" (without extension).
		userHomeDir, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(userHomeDir)
		viper.SetConfigName(".kubectl-karbon")
	}

	viper.SetEnvPrefix("karbon")
	viper.BindEnv("kubeconfig", "KUBECONFIG")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
