# Changelog

## v0.6.2

- **Packaging:**
  - Use new dragonchain version 4.3.3 (chart version 1.0.7)
  - Update default installed kubectl to v1.17.1
  - Update default installed virtualbox to v6.1.2
  - Use openfaas chart version 5.4.1 (openfaas 0.18.7, faas-netes 0.9.15)

## v0.6.1

- **Features:**
  - Add warning when running on arm64 that support is currently experimental and not yet fully working
- **Bugs:**
  - Fix a bug where installer would fail after inputting configuration if dragonchain config folder didn't already exist
  - Fix install script for arm64

## v0.6.0

- **Features:**
  - Use new dragonchain version 4.3.2
  - Add support for vmdriver=none with minikube on linux
  - Add partial-support for arm64 on linux (using vmdriver=none option) (experimental; not yet fully working due to [minikube support](https://github.com/kubernetes/minikube/issues/5667))
  - Use [local-path-provisioner](https://github.com/rancher/local-path-provisioner) for pvc storage
  - Add support to restart installation with previous installation configuration
- **Packaging:**
  - Use openfaas chart version 5.4.0
  - Use docker-registry chart version 1.9.1
  - Update default installed helm to v3.0.2
  - Update default installed kubectl to v1.17.0
  - Update default installed minikube to v1.6.2
  - Update default installed virtualbox to v6.1.0

## v0.5.1

- **Features:**
  - Use new Dragonchain version 4.3.0
- **Bugs:**
  - Make initializing helm more reliable

## v0.5.0

- **Features:**
  - Initial beta release. Future changelogs will have more information
