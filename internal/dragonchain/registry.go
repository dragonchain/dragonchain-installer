package dragonchain

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

var registryNamespacesYaml = []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: registry`)

func createDockerRegistryDeployment() error {
	// Create the necessary namespaces
	cmd := exec.Command("kubectl", "apply", "--context="+configuration.MinikubeContext, "-f", "-")
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewBuffer(registryNamespacesYaml)
	if err := cmd.Run(); err != nil {
		return errors.New("Error creating registry namespace:\n" + err.Error())
	}
	// Install the registry
	cmd = exec.Command("helm", "upgrade", "--install", "registry", "stable/docker-registry", "--namespace", "registry", "--set", "persistence.enabled=true,persistence.storageClass=local-path,persistence.deleteEnabled=true,service.type=ClusterIP,service.clusterIP="+configuration.RegistryIP+",service.port="+strconv.Itoa(configuration.RegistryPort), "--version", configuration.RegistryHelmVersion, "--kube-context", configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error helm deploying registry:\n" + err.Error())
	}
	return nil
}
