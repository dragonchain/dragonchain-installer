package dragonchain

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

func createDockerRegistryDeployment() error {
	// Install the registry
	cmd := exec.Command("helm", "upgrade", "--install", "registry", "stable/docker-registry", "--namespace", "registry", "--set", "persistence.enabled=true,persistence.storageClass=standard,persistence.deleteEnabled=true,service.type=ClusterIP,service.clusterIP="+configuration.RegistryIP+",service.port="+strconv.Itoa(configuration.RegistryPort), "--version", configuration.RegistryHelmVersion, "--kube-context", configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error helm deploying openfaas:\n" + err.Error())
	}
	return nil
}
