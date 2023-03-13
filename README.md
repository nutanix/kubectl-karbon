## kubectl-karbon - Quickly connect to your Karbon cluster!

This kubectl extension allows to quickly connect to an existing karbon cluster without the need to connect to Prism UI.
It uses the Karbon API to get kubeconfig file and install it on your local system.

## Demo

https://user-images.githubusercontent.com/180613/117446386-4d04f080-af3c-11eb-866d-282bc14ee97a.mp4

---

[![Go Report Card](https://goreportcard.com/badge/github.com/nutanix/kubectl-karbon)](https://goreportcard.com/report/github.com/nutanix/kubectl-karbon)
[![Maintainability](https://api.codeclimate.com/v1/badges/e0f43f5c74eabfa8bc4d/maintainability)](https://codeclimate.com/github/nutanix/kubectl-karbon/maintainability)
![CI](https://github.com/nutanix/kubectl-karbon/actions/workflows/ci.yml/badge.svg)
![Release](https://github.com/nutanix/kubectl-karbon/actions/workflows/release.yml/badge.svg)

[![release](https://img.shields.io/github/release-pre/nutanix/kubectl-karbon.svg)](https://github.com/nutanix/kubectl-karbon/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/nutanix/kubectl-karbon/blob/master/LICENSE)
![Proudly written in Golang](https://img.shields.io/badge/written%20in-Golang-92d1e7.svg)
[![Releases](https://img.shields.io/github/downloads/nutanix/kubectl-karbon/total.svg)](https://github.com/nutanix/kubectl-karbon/releases)

---



## Installation

There are several installation options:

- As kubectl Krew plugins
- As Homebrew or Linuxbrew package
- Manual installation

After installing, the tools will be available as `kubectl karbon`.

### Kubectl Plugins

You can install and use [Krew](https://github.com/kubernetes-sigs/krew/) kubectl
plugin manager to get the `karbon` plugin .

```sh
kubectl krew install karbon
```

Keep up-to-date with `kubectl krew upgrade karbon` (or kubectl krew upgrade to upgrade everything)

### Homebrew Package

You can install with [Homebrew](https://brew.sh) for macOS or [LinuxBrew](https://docs.brew.sh/Homebrew-on-Linux) for Linux

```sh
brew install nutanix/tap/kubectl-karbon
```

Keep up-to-date with `brew upgrade kubectl-karbon` (or brew upgrade to upgrade everything)

### Manual

 - Download your corresponding [release](https://github.com/nutanix/kubectl-karbon/releases)
 - Install the binary somewhere in your PATH (/usr/local/bin for example)
 - use it with `kubectl karbon`

***MacOS X notes for security error***

 Depending of your OS settings when you install you binary manually we must launch the following command:
 `xattr -r -d com.apple.quarantine /usr/local/bin/kubectl-karbon`

## Usage

* `kubectl karbon help` Help about any command
* `kubectl karbon list` Get the list of k8s clusters
* `kubectl karbon login` Authenticate user with Nutanix Prism Central, create kubeconfig file, get ssh key/cert, ...
* `kubectl karbon logout` Remove kubeconfig file, remove ssh key/cert file, clean ssh-agent ...
* `kubectl karbon version` Print the version of the plugin

### Config file

You can specify a config file to define your seetings. The default is $HOME/.kubectl-karbon.yaml and you can use another one with the `--config` flag.

```yaml
server: servername
port: 9440
cluster: karbon_cluster_name
user: admin
insecure: true
verbose: false
force: false
kubie: false
keyring: false
#ssh-agent: false
#ssh-file: false
#kubie-path: ~/.kube/.kubie/
#kubeconfig: /path/.kube/config
```
*config file example*

All entries are optional, you can define only what you need to enforce.

### Env variables

you can also use the following environement variable

`KARBON_SERVER`  
`KARBON_PORT`  
`KARBON_CLUSTER`  
`KARBON_USER`  
`KARBON_INSECURE`  
`KARBON_VERBOSE`  
`KARBON_FORCE`  
`KARBON_PASSWORD`  
`KARBON_KUBIE`  
`KARBON_KUBIE_PATH`  
`KARBON_SSH_AGENT`  
`KARBON_SSH_FILE`  
`KARBON_KEYRING`
`KUBECONFIG`

precedence is

`FLAGS` => `ENV` => `CONFIG FILE` => `DEFAULT`

## File overwrite

You can use the `--force` option to overwrite any existing file(s) like kubeconfig or ssh key/cert.

## Password

By default this tools never stored the password.  
You can use the `KARBON_PASSWORD` env variable.  
You can also use the `--keyring` option to save and retrieve your password from the system keyring. It supports OS X, Linux/BSD (dbus) and Windows.  
In all other cases password should be provided in an interactive way.

## SSH option

During login, allow SSH key and cert retrieval.  
The key and cert can be added to the running ssh-agent (`--ssh-agent`) or saved in file inside the ~/.ssh/ directory (`--ssh-file`).

## Kubie mode

Allows full integration with [Kube](https://github.com/funkolab/kube) or [Kubie](https://blog.sbstp.ca/introducing-kubie/) who have support for split configuration files, meaning it can load Kubernetes contexts from multiple files.  
When this mode is active (`--kubie`) each kubeconfig file is stored as an independent file in the kubie-path directoy (default `~/.kube/kubie/cluster_name.yaml`)

## Building From Source

 kubectl-karbon is currently using go v1.16 or above. In order to build  kubectl-karbon from source you must:

 1. Clone the repo
 2. Build and run the executable

      ```shell
      make build && make install
      ```
