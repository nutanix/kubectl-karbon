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
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/term"
)

type nutanixCluster struct {
	server   string
	login    string
	password string
	port     int
	timeout  int
	insecure bool
}

type karbonClusterList struct {
	KubeapiServerIpv4Address string `json:"kubeapi_server_ipv4_address"`
	Name                     string `json:"name"`
	Status                   string `json:"status"`
	UUID                     string `json:"uuid"`
	Version                  string `json:"version"`
}

type kubeConfig struct {
	KubeConfig string `json:"kube_config"`
}

type sshConfig struct {
	Certificate string `json:"certificate"`
	ExpiryTime  string `json:"expiry_time"`
	PrivateKey  string `json:"private_key"`
	Username    string `json:"username"`
}

func (nutanix *nutanixCluster) selectCluster() (string, error) {
	clusters, err := nutanix.listKarbonClusters()
	if err != nil {
		return "", err
	}

	clustersList := []string{}
	for _, cluster := range clusters {
		clustersList = append(clustersList, cluster.Name)
	}

	prompt := promptui.Select{
		Label: "Select a Cluster",
		Items: clustersList,
	}

	_, result, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return "", err
	}

	return result, nil
}

func (nutanix *nutanixCluster) listKarbonClusters() ([]karbonClusterList, error) {

	karbonListUrl := "/karbon/v1-beta.1/k8s/clusters"
	method := "GET"

	if verbose {
		fmt.Printf("Retrieve cluster list\n")
	}

	ResponseJSON, err := nutanix.clusterRequest(method, karbonListUrl, nil)
	cobra.CheckErr(err)

	var clusters []karbonClusterList

	err = json.Unmarshal([]byte(ResponseJSON), &clusters)
	if err != nil {
		return nil, err
	}

	return clusters, nil
}

func saveKeyFile(cluster string, ssh sshConfig, force bool) error {

	privateKey := []byte(ssh.PrivateKey)
	certificate := []byte(ssh.Certificate)

	userHomeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	sshDir := filepath.Join(userHomeDir, ".ssh")

	// Create the directory if it does not exist
	err = os.MkdirAll(sshDir, 0700)
	cobra.CheckErr(err)

	// Write the private key
	privateKeyFile := filepath.Join(sshDir, cluster)

	_, err = os.Stat(privateKeyFile)
	if err == nil && !force {
		return fmt.Errorf("file %s already exist, use force option to overwrite it", privateKeyFile)
	}

	err = ioutil.WriteFile(privateKeyFile, privateKey, 0600)
	cobra.CheckErr(err)

	// Write the certificate
	certificateFile := filepath.Join(sshDir, fmt.Sprintf("%s-cert.pub", cluster))

	_, err = os.Stat(certificateFile)
	if err == nil && !force {
		return fmt.Errorf("file %s already exist, use force option to overwrite it", certificateFile)
	}

	err = ioutil.WriteFile(certificateFile, certificate, 0600)
	cobra.CheckErr(err)

	if verbose {
		fmt.Printf("privateKey file %s successfully written\n", privateKeyFile)
		fmt.Printf("certificate file %s successfully written\n", certificateFile)
	}
	return nil

}

func deleteKeyFile(cluster string) error {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sshDir := filepath.Join(userHomeDir, ".ssh")

	privateKeyFile := filepath.Join(sshDir, cluster)
	err = os.Remove(privateKeyFile)
	if err != nil {
		return err
	}

	certificateFile := filepath.Join(sshDir, fmt.Sprintf("%s-cert.pub", cluster))
	err = os.Remove(certificateFile)
	if err != nil {
		return err
	}

	if verbose {
		fmt.Printf("privateKey file %s successfully deleted\n", privateKeyFile)
		fmt.Printf("certificate file %s successfully deleted\n", certificateFile)
	}

	return nil
}

func addKeyAgent(cluster string, ssh sshConfig) error {

	expiryTime := ssh.ExpiryTime
	privateKey := []byte(ssh.PrivateKey)
	certificate := []byte(ssh.Certificate)

	// Get the ssh agent
	socket := os.Getenv("SSH_AUTH_SOCK")

	if socket == "" {
		fmt.Println("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	cobra.CheckErr(err)

	agentClient := agent.NewClient(conn)

	data, _ := pem.Decode(privateKey)
	parsedKey, err := x509.ParsePKCS1PrivateKey(data.Bytes)
	cobra.CheckErr(err)

	sshCert, err := unmarshalCert(certificate)
	cobra.CheckErr(err)

	now := time.Now()
	layout := "2006-01-02T15:04:05.000Z"
	futureDate, err := time.Parse(layout, expiryTime)
	cobra.CheckErr(err)
	diff := futureDate.Sub(now)

	err = agentClient.Add(agent.AddedKey{
		PrivateKey:   parsedKey,
		Certificate:  sshCert,
		Comment:      fmt.Sprintf("karbon cluster %s", cluster),
		LifetimeSecs: uint32(diff.Seconds()),
	})
	cobra.CheckErr(err)

	if verbose {
		fmt.Printf("SSH key for cluster '%s' added to ssh-agent\n", cluster)
	}
	return nil

}

func deleteKeyAgent(cluster string) error {

	// Get the ssh agent
	socket := os.Getenv("SSH_AUTH_SOCK")

	if socket == "" {
		fmt.Println("SSH_AUTH_SOCK environment variable not set")
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return err
	}

	agentClient := agent.NewClient(conn)

	keyList, err := agentClient.List()
	if err != nil {
		return err
	}

	searchString := fmt.Sprintf("karbon cluster %s", cluster)

	for _, key := range keyList {
		if key.Comment == searchString {
			err = agentClient.Remove(key)
			if err != nil {
				return err
			}
			if verbose {
				fmt.Printf("SSH key for cluster '%s' deleted from ssh-agent\n", cluster)
			}
		}
	}

	return nil
}

func unmarshalCert(bytes []byte) (*ssh.Certificate, error) {
	pub, _, _, _, err := ssh.ParseAuthorizedKey(bytes)
	if err != nil {
		return nil, err
	}
	cert, ok := pub.(*ssh.Certificate)
	if !ok {
		return nil, fmt.Errorf("failed to cast to certificate")
	}
	return cert, nil
}

func getCredentials() (string, string) {
	userArg := viper.GetString("user")

	var password string
	var ok bool
	password, ok = os.LookupEnv("KARBON_PASSWORD")

	if !ok {
		fmt.Printf("Enter %s password:\n", userArg)
		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		cobra.CheckErr(err)

		password = string(bytePassword)
	}
	return userArg, password
}

func newNutanixCluster() (*nutanixCluster, error) {
	server := viper.GetString("server")
	if server == "" {
		return nil, fmt.Errorf("error: required flag \"server\" not set")
	}

	userArg, password := getCredentials()

	c := nutanixCluster{
		server:   server,
		login:    userArg,
		password: password,
		port:     viper.GetInt("port"),
		timeout:  viper.GetInt("timeout"),
		insecure: viper.GetBool("insecure"),
	}
	return &c, nil
}

func (c *nutanixCluster) clusterRequest(method string, path string, payload []byte) ([]byte, error) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: c.insecure}

	client := &http.Client{Transport: customTransport, Timeout: time.Second * time.Duration(c.timeout)}
	requestUrl := fmt.Sprintf("https://%s:%d/%s", c.server, c.port, path)
	req, err := http.NewRequest(method, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.login, c.password)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 401:
		return nil, fmt.Errorf("invalid client credentials")
	case 404:
		return nil, fmt.Errorf("karbon cluster not found")
	case 200:
		// OK
	default:
		return nil, fmt.Errorf("internal Error")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
