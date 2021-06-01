

## kubectl-karbon - Quickly connect to your Karbon cluster!

This kubectl extension allow to quickly connect to an existing karbon cluster without the need to connect to Prism UI.
He use the Karbon API to get kubeconfig file and install it on your local system.

## Demo

https://user-images.githubusercontent.com/180613/117446386-4d04f080-af3c-11eb-866d-282bc14ee97a.mp4

---

 [![Go Report Card](https://goreportcard.com/badge/github.com/nutanix/kubectl-karbon)](https://goreportcard.com/report/github.com/nutanix/kubectl-karbon)
![CI](https://github.com/nutanix/kubectl-karbon/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/nutanix/kubectl-karbon/actions/workflows/release.yml/badge.svg)

[![release](https://img.shields.io/github/release-pre/nutanix/kubectl-karbon.svg)](https://github.com/nutanix/kubectl-karbon/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/mum4k/termdash/blob/master/LICENSE)
[![Releases](https://img.shields.io/github/downloads/nutanix/kubectl-karbon/total.svg)](https://github.com/nutanix/kubectl-karbon/releases)

---



## Install

 - Download your corresponding [release](https://github.com/nutanix/kubectl-karbon/releases)
 - Install the binary somewhere in your PATH (/usr/local/bin for example)
 - use it with `kubectl karbon`

***MacOS X notes for security error***

 once your binary installed launch the following command:
 `xattr -r -d com.apple.quarantine /usr/local/bin/kubectl-karbon`

## Usage

`kubectl karbon help`

### config file

You can specify a config file to define your seetings. The default is $HOME/.kubectl-karbon.yaml and you can use another one with the `--config` flag.

```yaml
server: servername
port: 9440
cluster: karbon_cluster_name
user: admin
insecure: true
verbose: false
kubeconfig: /path/.kube/config
```
*config file example*

All entries are optional, you can define only what you need to enforce.

### env variables

you can also use the following environement variable

`KARBON_SERVER`  
`KARBON_PORT`  
`KARBON_CLUSTER`  
`KARBON_USER`  
`KARBON_INSECURE`  
`KARBON_VERBOSE`  
`KUBECONFIG`

precedence is

`FLAGS` => `ENV` => `CONFIG FILE` => `DEFAULT`

## Building From Source

 kubectl-karbon is currently using go v1.16 or above. In order to build  kubectl-karbon from source you must:

 1. Clone the repo
 2. Build and run the executable

      ```shell
      make build && make install
      ```
