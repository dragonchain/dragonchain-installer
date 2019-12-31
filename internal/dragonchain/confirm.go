package dragonchain

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
		Metadata (struct {
			Name string `json:"name"`
		}) `json:"metadata"`
		Status (struct {
			Phase             string `json:"phase"`
			ContainerStatuses [](struct {
				Ready bool `json:"ready"`
			}) `json:"containerStatuses"`
		}) `json:"status"`
	}) `json:"items"`
}

// GetDragonchainPublicID gets the public id of a running dragonchain
func GetDragonchainPublicID(config *configuration.Configuration) (string, error) {
	return dragonchainPubIDRecurse(config, 0)
}

func dragonchainPubIDRecurse(config *configuration.Configuration, tries int) (string, error) {
	if tries > 60 {
		return "", errors.New("Too long waiting for running dragonchain pod. Check kubernetes cluster for more information")
	}
	// Get the webserver pod which we can exec into
	cmd := exec.Command("kubectl", "get", "pod", "-n", "dragonchain", "-l", "app.kubernetes.io/component=webserver,dragonchainId="+config.InternalID, "-o", "json", "--context="+configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("Error checking for dragonchain pods:\n" + err.Error())
	}
	var chainList kubectlPodJSONList
	if err := json.Unmarshal(output, &chainList); err != nil {
		return "", errors.New("Failed to parse pod list from kubectl:\n" + err.Error())
	}
	// Make sure pod is running before using it
	if len(chainList.Items) < 1 || chainList.Items[0].Status.Phase != "Running" {
		// If no pod is ready, wait and recurse to try again
		time.Sleep(1 * time.Second)
		return dragonchainPubIDRecurse(config, tries+1)
	}
	// Exec into the pod with the command to get the chain's public id
	cmd = exec.Command("kubectl", "exec", "-n", "dragonchain", chainList.Items[0].Metadata.Name, "--context="+configuration.MinikubeContext, "--", "python3", "-c", "from dragonchain.lib.keys import get_public_id; print(get_public_id())")
	cmd.Stderr = os.Stderr
	output, err = cmd.Output()
	if err != nil {
		return "", errors.New("Error executing in dragonchain pod " + chainList.Items[0].Metadata.Name + ":\n" + err.Error())
	}
	outputStr := string(output)
	outputStr = strings.TrimSuffix(outputStr, "\r\n")
	outputStr = strings.TrimSuffix(outputStr, "\n")
	return outputStr, nil
}

func waitForDragonchainToBeReady(config *configuration.Configuration) error {
	// Wait up to 120-ish seconds for dragonchain to be ready before erroring
	for i := 0; i < 120; i++ {
		// Wait before checking
		time.Sleep(1 * time.Second)
		// Print a '.' every 10 seconds to show that the program is still running
		if i%10 == 0 {
			fmt.Print(".")
		}
		cmd := exec.Command("kubectl", "get", "pod", "-n", "dragonchain", "-l", "dragonchainId="+config.InternalID, "-o", "json", "--context="+configuration.MinikubeContext)
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			return errors.New("Error checking dragonchain pods:\n" + err.Error())
		}
		var podList kubectlPodJSONList
		if err := json.Unmarshal(output, &podList); err != nil {
			return errors.New("Failed to parse pod list from kubectl:\n" + err.Error())
		}
		ready := true
		for _, items := range podList.Items {
			ready = ready && items.Status.Phase == "Running" // Only ready if all containers are "Running"
			for _, status := range items.Status.ContainerStatuses {
				ready = ready && status.Ready // Only ready if all containers are also 'ready'
			}
		}
		if ready {
			return nil
		}
	}
	return errors.New("Dragonchain pods failed to become ready. Check kubernetes cluster for more information")
}
