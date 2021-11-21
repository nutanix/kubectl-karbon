/*
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
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var sshResponseJSON map[string]interface{}

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Get SSH credentials to access to the k8s cluster",
	Long:  `Get SSH credentials to remotely access nodes belonging to the k8s cluster. The credentials have an expiry time of 24 hours.`,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("server", cmd.Flags().Lookup("server"))
		viper.BindPFlag("cluster", cmd.Flags().Lookup("cluster"))
		viper.BindPFlag("user", cmd.Flags().Lookup("user"))
		viper.BindPFlag("port", cmd.Flags().Lookup("port"))
		viper.BindPFlag("insecure", cmd.Flags().Lookup("insecure"))
		viper.BindPFlag("agent", cmd.Flags().Lookup("agent"))
		viper.BindPFlag("file", cmd.Flags().Lookup("file"))
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

		karbonSSHUrl := fmt.Sprintf("https://%s:%d/karbon/v1/k8s/clusters/%s/ssh", server, port, cluster)
		method := "GET"

		if verbose {
			fmt.Printf("Connect on https://%s:%d/ and retrieve SSH credentials for cluster %s\n", server, port, cluster)
		}

		insecureSkipVerify := viper.GetBool("insecure")

		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}

		timeout, _ := cmd.Flags().GetInt("request-timeout")
		client := &http.Client{Transport: customTransport, Timeout: time.Second * time.Duration(timeout)}
		req, err := http.NewRequest(method, karbonSSHUrl, nil)
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

		json.Unmarshal([]byte(body), &sshResponseJSON)

		if viper.GetBool("file") {
			saveKeyFile(cluster, sshResponseJSON)
		}

		if viper.GetBool("agent") {
			addKeyAgent(cluster, sshResponseJSON)
		}
	},
}

func init() {
	rootCmd.AddCommand(sshCmd)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	sshCmd.Flags().String("server", "", "Address of the PC to authenticate against")

	sshCmd.Flags().StringP("user", "u", user.Username, "Username to authenticate")

	sshCmd.Flags().String("cluster", "", "Karbon cluster to connect against")

	sshCmd.Flags().Int("port", 9440, "Port to run Application server on")

	sshCmd.Flags().BoolP("insecure", "k", false, "Skip certificate verification (this is insecure)")

	sshCmd.Flags().Bool("agent", true, "Add Key and Cert in SSH agent")
	sshCmd.Flags().Bool("file", false, "Store Key and Cert in local directory (default ~/.ssh/)")
}
