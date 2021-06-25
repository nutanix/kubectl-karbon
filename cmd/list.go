/*
Package cmd list existing karbon cluster(s)
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
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Get the list of k8s clusters",
	Long:  `Return the list of all kubernetes cluster running on the tergeted Nutanix Karbon platform`,
	PreRun: func(cmd *cobra.Command, args []string) {

		viper.BindPFlag("server", cmd.Flags().Lookup("server"))
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

		port := viper.GetInt("port")

		karbonListUrl := fmt.Sprintf("https://%s:%d/karbon/v1-beta.1/k8s/clusters", server, port)
		method := "GET"

		if verbose {
			fmt.Printf("Connect on https://%s:%d/ and retrieve cluster list\n", server, port)
		}

		insecureSkipVerify := viper.GetBool("insecure")

		customTransport := http.DefaultTransport.(*http.Transport).Clone()
		customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify}

		timeout, _ := cmd.Flags().GetInt("request-timeout")
		client := &http.Client{Transport: customTransport, Timeout: time.Second * time.Duration(timeout)}
		req, err := http.NewRequest(method, karbonListUrl, nil)
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
		case 200:
			// OK
		default:
			fmt.Println("Internal Error")
			return

		}

		body, err := ioutil.ReadAll(res.Body)
		cobra.CheckErr(err)

		clusters := make([]map[string]interface{}, 0)
		json.Unmarshal([]byte(body), &clusters)

		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)

		defer w.Flush()

		fmt.Fprintf(w, "%s\t%s\t%s\t", "NAME", "VERSION", "STATUS")

		for _, cluster := range clusters {
			fmt.Fprintf(w, "\n%s\tv%s\t%s\t", cluster["name"], cluster["version"], cluster["status"].(string)[1:])
		}
		fmt.Fprintf(w, "\n")

	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	listCmd.Flags().String("server", "", "Address of the PC to authenticate against")

	listCmd.Flags().StringP("user", "u", user.Username, "Username to authenticate")

	listCmd.Flags().Int("port", 9440, "Port to run Application server on")

	listCmd.Flags().BoolP("insecure", "k", false, "Skip certificate verification (this is insecure)")
}
