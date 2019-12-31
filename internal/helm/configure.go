package helm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

type kubectlPodJSONList struct {
	Items [](struct {
		Status (struct {
			ContainerStatuses [](struct {
				Ready bool `json:"ready"`
			}) `json:"containerStatuses"`
		}) `json:"status"`
	}) `json:"items"`
}

func waitForTillerToBeReady() error {
	// Wait up to 60 seconds-ish for tiller pod to be ready before erroring
	for i := 0; i < 60; i++ {
		// Wait before checking
		time.Sleep(1 * time.Second)
		cmd := exec.Command("kubectl", "get", "pod", "-n", "kube-system", "-l", "name=tiller", "-o", "json", "--context="+configuration.MinikubeContext)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return errors.New("Error checking for tiller pod " + err.Error())
		}
		var podList kubectlPodJSONList
		if err := json.Unmarshal(output, &podList); err != nil {
			return errors.New("Failed to parse pod list from kubectl:\n" + err.Error())
		}
		for _, items := range podList.Items {
			for _, status := range items.Status.ContainerStatuses {
				if status.Ready {
					return nil
				}
			}
		}
	}
	return errors.New("Tiller pod failed to become ready")
}

// GetHelmMajorVersion gets the major version of helm (either 2 or 3)
func GetHelmMajorVersion() (int, error) {
	cmd := exec.Command("helm", "version", "-c", "--short")
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return 0, errors.New("Unable to get helm version:\n" + err.Error())
	}
	versionOutput := string(out)
	if strings.Contains(versionOutput, "v2.") {
		return 2, nil
	} else if strings.Contains(versionOutput, "v3.") {
		return 3, nil
	}
	return 0, errors.New("Unable to parse helm version string")
}

// InitializeHelm intializes helm both locally, and in the minikube cluster
func InitializeHelm() error {
	fmt.Println("Configuring helm")
	helmVersion, err := GetHelmMajorVersion()
	if err != nil {
		return err
	}
	// Only helm v2 requires tiller initialization
	if helmVersion == 2 {
		cmd := exec.Command("helm", "init", "--upgrade", "--kube-context", configuration.MinikubeContext)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.New("Initializing helm failed:\n" + err.Error())
		}
	}
	cmd := exec.Command("helm", "repo", "add", "dragonchain", "https://dragonchain-charts.s3.amazonaws.com")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Adding dragonchain helm repo failed:\n" + err.Error())
	}
	cmd = exec.Command("helm", "repo", "add", "openfaas", "https://openfaas.github.io/faas-netes/")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Adding openfaas helm repo failed:\n" + err.Error())
	}
	if helmVersion >= 3 {
		// Stable repository is not added by default in helm 3+
		cmd = exec.Command("helm", "repo", "add", "stable", "https://kubernetes-charts.storage.googleapis.com")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return errors.New("Adding stable helm repo failed:\n" + err.Error())
		}
	}
	cmd = exec.Command("helm", "repo", "update")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Updating helm repo failed (are you connected to the internet?):\n" + err.Error())
	}
	if helmVersion == 2 {
		time.Sleep(3 * time.Second)
		if err := waitForTillerToBeReady(); err != nil {
			return err
		}
	}
	return nil
}
