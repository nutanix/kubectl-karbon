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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// flags variable

var kubeconfigResponseJSON map[string]interface{}

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate user with Nutanix Prism Central",
	Long:  `Authenticate user with Nutanix Prism Central and create a local kubeconfig file for the selected cluster`,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("server", cmd.Flags().Lookup("server"))
		viper.BindPFlag("cluster", cmd.Flags().Lookup("cluster"))
		viper.BindPFlag("user", cmd.Flags().Lookup("user"))
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("insecure", cmd.Flags().Lookup("insecure"))
	},
	Run: func(cmd *cobra.Command, args []string) {

		server := viper.GetString("server")
		if server == "" {
			fmt.Fprintln(os.Stderr, "Error: required flag(s) \"server\" not set")
			cmd.Usage()
			return
		}

		cluster := viper.GetString("cluster")
		if cluster == "" {
			fmt.Fprintln(os.Stderr, "Error: required flag(s) \"cluster\" not set")
			cmd.Usage()
			return
		}

		port := viper.GetInt("port")

		karbonKubeconfigUrl := fmt.Sprintf("https://%s:%d/karbon/v1/k8s/clusters/%s/kubeconfig", server, port, cluster)
		method := "GET"

		if verbose {
			fmt.Printf("Connect on https://%s:%d/ and retrieve Kubeconfig for cluster %s\n", server, port, cluster)
		}

		insecureSkipVerify := viper.GetBool("insecure")

		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}

		timeout, _ := cmd.Flags().GetInt("request-timeout")
		client := &http.Client{Transport: customTransport, Timeout: time.Second * time.Duration(timeout)}
		req, err := http.NewRequest(method, karbonKubeconfigUrl, nil)
		cobra.CheckErr(err)

		userArg, password := getCredentials()

		req.SetBasicAuth(userArg, password)

		res, err := client.Do(req)
		cobra.CheckErr(err)

		defer res.Body.Close()

		switch res.StatusCode {
		case 401:
			fmt.Println("Invalid client credentials")
			return
		case 404:
			fmt.Printf("K8s cluster %s not found\n", cluster)
			return
		case 200:
			// OK
		default:
			fmt.Println("Internal Error")
			return

		}

		body, err := ioutil.ReadAll(res.Body)
		cobra.CheckErr(err)

		// fmt.Println(string(body))
		json.Unmarshal([]byte(body), &kubeconfigResponseJSON)
		// fmt.Printf(kubeconfigResponseJSON["kube_config"].(string))

		data := []byte(kubeconfigResponseJSON["kube_config"].(string))

		kubeconfig := viper.GetString("kubeconfig")

		if verbose {
			fmt.Printf("Kubeconfig file %s successfully written\n", kubeconfig)
		}

		kubeconfigPath := filepath.Dir(kubeconfig)
		_, err = os.Stat(kubeconfigPath)

		if os.IsNotExist(err) {
			err := os.MkdirAll(kubeconfigPath, 0700)
			cobra.CheckErr(err)
		}

		err = ioutil.WriteFile(kubeconfig, data, 0600)
		cobra.CheckErr(err)

		fmt.Printf(`Logged in successfully

You have access to the following cluster:
  %s
        `, cluster)

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
}
