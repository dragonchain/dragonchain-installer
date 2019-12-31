package dragonchain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
	"github.com/dragonchain/dragonchain-installer/internal/helm"
)

func doesHelmDeploymentExist(name string, namespace string) (bool, error) {
	helmVersion, err := helm.GetHelmMajorVersion()
	if err != nil {
		return false, err
	}
	cmd := exec.Command("helm", "get", "notes", name, "--kube-context", configuration.MinikubeContext)
	if helmVersion > 2 {
		cmd = exec.Command("helm", "get", "notes", name, "-n", namespace, "--kube-context", configuration.MinikubeContext)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Will get an error if the deployment does not exist
		return false, nil
	}
	// If command ran successfully, a helm installation already exists for this deployment
	return true, nil
}

// SetupDragonchainPreReqs sets up kubernetes resource requirements for dragonchain
func SetupDragonchainPreReqs(config *configuration.Configuration) error {
	if err := exec.Command("kubectl", "apply", "-f", "https://raw.githubusercontent.com/rancher/local-path-provisioner/master/deploy/local-path-storage.yaml").Run(); err != nil {
		return errors.New("Error creating local path provisioner:\n" + err.Error())
	}
	if exec.Command("kubectl", "get", "namespace", "dragonchain", "--context="+configuration.MinikubeContext).Run() != nil {
		// Create the dragonchain namespace if necessary
		fmt.Println("Creating dragonchain namespace")
		cmd := exec.Command("kubectl", "create", "namespace", "dragonchain", "--context="+configuration.MinikubeContext)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.New("Error creating dragonchain namespace:\n" + err.Error())
		}
	}
	// Set up l1 dependencies if needed
	if config.Level == 1 {
		// Set up openfaas
		exists, err := doesHelmDeploymentExist("openfaas", "openfaas")
		if err != nil {
			return errors.New("Error checking for existing openfaas installation:\n" + err.Error())
		}
		if !exists {
			fmt.Println("Openfaas does not appear to be installed. Installing now")
			if err := createOpenFaasDeployment(); err != nil {
				return err
			}
		}
		if !config.UseVM {
			// Try to backup old docker daemon config if it exists
			cmd := exec.Command("sudo", "mv", "/etc/docker/daemon.json", "/etc/docker/daemon.json.bak")
			cmd.Stdin = os.Stdin
			cmd.Run()
			// If using native machine docker, need to ensure that insecure registry for the registry is set on the daemon
			dockerDaemonJSON := "{\\\"insecure-registries\\\":[\\\"" + configuration.RegistryIP + ":" + strconv.Itoa(configuration.RegistryPort) + "\\\"]}"
			cmd = exec.Command("sh", "-c", "echo "+dockerDaemonJSON+" | sudo tee /etc/docker/daemon.json")
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return errors.New("Error setting insecure registry setting with docker daemon:\n" + err.Error())
			}
			cmd = exec.Command("sudo", "service", "docker", "restart")
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return errors.New("Error restarting docker daemon:\n" + err.Error())
			}
			// Briefly wait for containers to come back up after restarting
			time.Sleep(10 * time.Second)
		}
		// Set up docker registry
		exists, err = doesHelmDeploymentExist("registry", "registry")
		if err != nil {
			return errors.New("Error checking for existing container registry installation:\n" + err.Error())
		}
		if !exists {
			fmt.Println("Docker registry does not appear to be installed. Installing now")
			if err := createDockerRegistryDeployment(); err != nil {
				return err
			}
		}
		// Set up openfaas builder service account
		if !openfaasServiceAccountExists() {
			fmt.Println("Openfaas builder service account doesn't exist. Creating now")
			if err := createOpenfaasBuilderServiceAccount(); err != nil {
				return err
			}
		}
	}
	return nil
}
