package minikube

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

type minikubeProfileList struct {
	Valid [](struct {
		Name string `json:"Name"`
	}) `json:"valid"`
}

func existingMinikubeClusterExists() (bool, error) {
	// Make sure minikube profiles folder exists or minikube can unexpectedly fail: https://github.com/kubernetes/minikube/issues/5898
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, errors.New("Error getting home dir:\n" + err.Error())
	}
	if err := os.MkdirAll(filepath.Join(homeDir, ".minikube", "profiles"), os.ModePerm); err != nil {
		return false, errors.New("Failed to confirm or create minikube profiles folder:\n" + err.Error())
	}
	// Get profile list from minikube
	cmd := exec.Command("minikube", "profile", "list", "-o", "json")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return false, errors.New("Couldn't get minikube profile list:\n" + err.Error())
	}
	var profileList minikubeProfileList
	if err := json.Unmarshal(out, &profileList); err != nil {
		return false, errors.New("Failed to parse profile list from minikube:\n" + err.Error())
	}
	for _, value := range profileList.Valid {
		if value.Name == configuration.MinikubeContext {
			return true, nil
		}
	}
	return false, nil
}

// FriendlyStartStopCommand returns the strings of the start/stop commands that a user can use to start stop minikube (and thus the dragonchain)
func FriendlyStartStopCommand() (startCommand string, stopCommand string) {
	startCommand = "minikube start -p " + configuration.MinikubeContext + " --kubernetes-version=" + configuration.KubernetesVersion
	stopCommand = "minikube stop -p " + configuration.MinikubeContext
	return
}

// StartMinikubeCluster starts (or creates and starts) the minikube cluster with a configured profile
func StartMinikubeCluster() error {
	// Switch current directory to the systemroot on C:\ if running on windows to avoid minikube bug: https://github.com/kubernetes/minikube/issues/1574
	if configuration.Windows {
		systemRoot, exists := os.LookupEnv("SYSTEMROOT")
		if !exists {
			return errors.New("Environment variable 'SYSTEMROOT' does not exist")
		}
		if err := os.Chdir(systemRoot); err != nil {
			return errors.New("Error switching directory:\n" + err.Error())
		}
	}
	exists, err := existingMinikubeClusterExists()
	if err != nil {
		return err
	}
	os.Setenv("MINIKUBE_IN_STYLE", "false")
	var minikubeStartCmd *exec.Cmd
	if exists {
		fmt.Println("\nStarting existing minikube cluster '" + configuration.MinikubeContext + "'; This can take a while")
		minikubeStartCmd = exec.Command("minikube", "start", "-p", configuration.MinikubeContext, "--kubernetes-version="+configuration.KubernetesVersion)
	} else {
		fmt.Println("\nStarting new minikube cluster '" + configuration.MinikubeContext + "'; This can take a while")
		minikubeStartCmd = exec.Command("minikube", "start", "-p", configuration.MinikubeContext, "--kubernetes-version="+configuration.KubernetesVersion, "--vm-driver=virtualbox", "--memory="+configuration.MinikubeVMMemory, "--cpus="+strconv.Itoa(configuration.MinikubeCpus))
	}
	minikubeStartCmd.Stdout = os.Stdout
	minikubeStartCmd.Stderr = os.Stderr
	if err := minikubeStartCmd.Run(); err != nil {
		return errors.New("Failed to start minikube. Resolve errors to continue:\n" + err.Error())
	}
	return nil
}
