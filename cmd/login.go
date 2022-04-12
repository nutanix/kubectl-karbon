/*
Package cmd login to the karbon cluster
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
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate user with Nutanix Prism Central",
	Long: `Authenticate user with Nutanix Prism Central and create a local kubeconfig file for the selected cluster.

If option enabled retrieve SSH key/cert and add them to ssh-agent or in file in ~/.ssh/ folder`,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("server", cmd.Flags().Lookup("server"))
		viper.BindPFlag("cluster", cmd.Flags().Lookup("cluster"))
		viper.BindPFlag("user", cmd.Flags().Lookup("user"))
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("insecure", cmd.Flags().Lookup("insecure"))
		viper.BindPFlag("kubie", cmd.Flags().Lookup("kubie"))
		viper.BindPFlag("kubie-path", cmd.Flags().Lookup("kubie-path"))
		viper.BindPFlag("ssh-agent", cmd.Flags().Lookup("ssh-agent"))
		viper.BindPFlag("ssh-file", cmd.Flags().Lookup("ssh-file"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		karbonCluster := viper.GetString("cluster")
		var err error

		if karbonCluster == "" {
			// fmt.Fprintln(os.Stderr, "Error: required flag \"cluster\" not set")
			// cmd.Usage()
			// return
			karbonCluster, err = selectCluster(cmd)
			cobra.CheckErr(err)
		}

		nutanixCluster, err := newNutanixCluster()
		if err != nil {
			fmt.Println(err)
			cmd.Usage()
			return
		}

		//  Kubeconfig management section

		if verbose {
			fmt.Printf("Connect on https://%s:%d/ and retrieve Kubeconfig for cluster %s\n", nutanixCluster.server, nutanixCluster.port, karbonCluster)
		}

		karbonKubeconfigPath := fmt.Sprintf("/karbon/v1/k8s/clusters/%s/kubeconfig", karbonCluster)
		method := "GET"

		kubeconfigResponseJSON, err := nutanixClusterRequest(nutanixCluster, method, karbonKubeconfigPath, nil)
		cobra.CheckErr(err)

		data := []byte(kubeconfigResponseJSON["kube_config"].(string))

		kubeconfig := viper.GetString("kubeconfig")

		if viper.GetBool("kubie") {
			kubiePath := viper.GetString("kubie-path")
			clusterFile := fmt.Sprintf("%s.yaml", karbonCluster)
			kubeconfig = filepath.Join(kubiePath, clusterFile)
		}

		if strings.HasPrefix(kubeconfig, "~/") {
			userHomeDir, err := os.UserHomeDir()
			cobra.CheckErr(err)
			kubeconfig = filepath.Join(userHomeDir, kubeconfig[2:])
		}

		kubeconfigPath := filepath.Dir(kubeconfig)
		_, err = os.Stat(kubeconfigPath)

		if os.IsNotExist(err) {
			err := os.MkdirAll(kubeconfigPath, 0700)
			cobra.CheckErr(err)
		}

		err = ioutil.WriteFile(kubeconfig, data, 0600)
		cobra.CheckErr(err)

		if verbose {
			fmt.Printf("Kubeconfig file %s successfully written\n", kubeconfig)
		}

		// SSH key/cert management section

		karbonSSHPath := fmt.Sprintf("/karbon/v1/k8s/clusters/%s/ssh", karbonCluster)
		method = "GET"

		if verbose {
			fmt.Printf("Connect on https://%s:%d/ and retrieve SSH key/cert for cluster %s\n", nutanixCluster.server, nutanixCluster.port, karbonCluster)
		}

		karbonSSHJSON, err := nutanixClusterRequest(nutanixCluster, method, karbonSSHPath, nil)
		cobra.CheckErr(err)

		if viper.GetBool("ssh-file") {
			err = saveKeyFile(karbonCluster, karbonSSHJSON)
			cobra.CheckErr(err)
		}

		if viper.GetBool("ssh-agent") {
			err = addKeyAgent(karbonCluster, karbonSSHJSON)
			cobra.CheckErr(err)
		}

		fmt.Printf("Logged successfully into %s cluster\n", karbonCluster)

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	loginCmd.Flags().String("server", "", "Address of the PC to authenticate against")

	loginCmd.Flags().StringP("user", "u", user.Username, "Username to authenticate")

	loginCmd.Flags().String("cluster", "", "Karbon cluster to connect against")

	loginCmd.Flags().Int("port", 9440, "Port to run Application server on")

	loginCmd.Flags().BoolP("insecure", "k", false, "Skip certificate verification (this is insecure)")

	loginCmd.Flags().Bool("kubie", false, "Store kubeconfig in independent file in kubie-path directory")

	userHomeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)
	defaultKubiePath := fmt.Sprintf("%s/.kube/kubie/", userHomeDir)
	loginCmd.Flags().String("kubie-path", defaultKubiePath, "Path to kubie kubeconfig directory")

	loginCmd.Flags().Bool("ssh-agent", false, "Add Key and Cert in SSH agent")
	loginCmd.Flags().Bool("ssh-file", false, "Save Key and Cert in ~/.ssh/ directory")
}
