package dragonchain

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

func doesHelmDeploymentExist(name string) (bool, error) {
	cmd := exec.Command("helm", "list", name, "--kube-context", configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return false, errors.New("Error checking helm for " + name + ":\n" + err.Error())
	}
	// If there is output, a helm installation already exists for this deployment
	return len(output) > 0, nil
}

// SetupDragonchainPreReqs sets up kubernetes resource requirements for dragonchain
func SetupDragonchainPreReqs(config *configuration.Configuration) error {
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
		exists, err := doesHelmDeploymentExist("openfaas")
		if err != nil {
			return errors.New("Error checking for existing openfaas installation:\n" + err.Error())
		}
		if !exists {
			fmt.Println("Openfaas does not appear to be installed. Installing now")
			if err := createOpenFaasDeployment(); err != nil {
				return err
			}
		}
		// Set up docker registry
		exists, err = doesHelmDeploymentExist("registry")
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
