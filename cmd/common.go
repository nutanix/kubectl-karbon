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
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func saveKeyFile(cluster string, sshResponseJSON map[string]interface{}) error {

	privateKey := []byte(sshResponseJSON["private_key"].(string))
	certificate := []byte(sshResponseJSON["certificate"].(string))

	userHomeDir, err := os.UserHomeDir()
	cobra.CheckErr(err)

	sshDir := filepath.Join(userHomeDir, ".ssh")

	// Create the directory if it does not exist
	err = os.MkdirAll(sshDir, 0700)
	cobra.CheckErr(err)

	// Write the private key
	privateKeyFile := filepath.Join(sshDir, cluster)
	err = ioutil.WriteFile(privateKeyFile, privateKey, 0600)
	cobra.CheckErr(err)

	// Write the certificate
	certificateFile := filepath.Join(sshDir, fmt.Sprintf("%s-cert.pub", cluster))
	err = ioutil.WriteFile(certificateFile, certificate, 0600)
	cobra.CheckErr(err)

	if verbose {
		fmt.Printf("privateKey file %s successfully written\n", privateKeyFile)
		fmt.Printf("certificate file %s successfully written\n", certificateFile)
	}
	return nil

}

func addKeyAgent(cluster string, sshResponseJSON map[string]interface{}) error {

	expiryTime := sshResponseJSON["expiry_time"].(string)
	privateKey := []byte(sshResponseJSON["private_key"].(string))
	certificate := []byte(sshResponseJSON["certificate"].(string))

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
		Comment:      fmt.Sprintf("karbon cluser %s", cluster),
		LifetimeSecs: uint32(diff.Seconds()),
	})
	cobra.CheckErr(err)

	if verbose {
		fmt.Printf("SSH key for cluster '%s' added to ssh-agent\n", cluster)
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
