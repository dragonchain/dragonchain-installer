package helm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

// InitializeHelm intializes helm both locally, and in the minikube cluster
func InitializeHelm() error {
	fmt.Println("Configuring helm")
	cmd := exec.Command("helm", "init", "--upgrade", "--kube-context", configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Initializing helm failed:\n" + err.Error())
	}
	cmd = exec.Command("helm", "repo", "add", "dragonchain", "https://dragonchain-charts.s3.amazonaws.com")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Adding dragonchain helm repo failed:\n" + err.Error())
	}
	cmd = exec.Command("helm", "repo", "add", "openfaas", "https://openfaas.github.io/faas-netes/")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Adding openfaas helm repo failed:\n" + err.Error())
	}
	cmd = exec.Command("helm", "repo", "update")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Updating helm repo failed (are you connected to the internet?):\n" + err.Error())
	}
	time.Sleep(3 * time.Second)
	if err := waitForTillerToBeReady(); err != nil {
		return err
	}
	return nil
}
