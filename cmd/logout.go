/*
Package cmd logout from the karbon cluster by removing kubeconfig file
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
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove all authentication items for the selected Karbon cluster",
	Long: `Remove all authentication items for the selected Karbon cluster.
	
Remove the local kubeconfig file, The SSH key/cert from file and SSH agent`,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("cluster", cmd.Flags().Lookup("cluster"))
		viper.BindPFlag("kubie", cmd.Flags().Lookup("kubie"))
		viper.BindPFlag("kubie-path", cmd.Flags().Lookup("kubie-path"))
		viper.BindPFlag("ssh-agent", cmd.Flags().Lookup("ssh-agent"))
		viper.BindPFlag("ssh-file", cmd.Flags().Lookup("ssh-file"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		karbonCluster := viper.GetString("cluster")
		if karbonCluster == "" {
			fmt.Fprintln(os.Stderr, "Error: required flag \"cluster\" not set")
			cmd.Usage()
			return
		}

		kubeconfig := viper.GetString("kubeconfig")

		if viper.GetBool("kubie") {
			kubiePath := viper.GetString("kubie-path")
			clusterFile := fmt.Sprintf("%s.yaml", karbonCluster)
			kubeconfig = filepath.Join(kubiePath, clusterFile)
		}

		err := os.Remove(kubeconfig)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		if viper.GetBool("ssh-file") {
			err := deleteKeyFile(karbonCluster)
			cobra.CheckErr(err)
		}

		if viper.GetBool("ssh-agent") {
			err = deleteKeyAgent(karbonCluster)
			cobra.CheckErr(err)
		}

		fmt.Printf("Logged out successfully from %s cluster\n", karbonCluster)
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)

	logoutCmd.Flags().String("cluster", "", "Karbon cluster to disconnect against")
	logoutCmd.Flags().Bool("kubie", false, "Remove kubeconfig independent file from kubie-path directory")

	userHomeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	defaultKubiePath := fmt.Sprintf("%s/.kube/kubie/", userHomeDir)
	logoutCmd.Flags().String("kubie-path", defaultKubiePath, "Path to kubie kubeconfig directory")

	logoutCmd.Flags().Bool("ssh-agent", false, "Remove Key and Cert from SSH agent")
	logoutCmd.Flags().Bool("ssh-file", false, "Remove Key and Cert from~/.ssh/ directory")
}
