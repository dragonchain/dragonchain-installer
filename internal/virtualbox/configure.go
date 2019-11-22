package virtualbox

import (
	"errors"
	"os"
	"os/exec"
	"strconv"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

func forwardVirtualboxPort(config *configuration.Configuration) error {
	// Delete possible existing port-forward rule before creating it (don't care about errors)
	exec.Command(vboxManageExecutable(), "controlvm", configuration.MinikubeContext, "natpf1", "delete", config.InternalID+"-traffic").Run()
	portStr := strconv.Itoa(config.Port)
	// Add host port-forwarding from VM network to host machine's network
	cmd := exec.Command(vboxManageExecutable(), "controlvm", configuration.MinikubeContext, "natpf1", config.InternalID+"-traffic,tcp,,"+portStr+",,"+portStr)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error forwarding virtualbox port (maybe this port is already in use on this machine?):\n" + err.Error())
	}
	return nil
}

// ConfigureVirtualboxVM configures the minikube virtualbox VM as necessary for dragonchain usage
func ConfigureVirtualboxVM(config *configuration.Configuration) error {
	if err := forwardVirtualboxPort(config); err != nil {
		return err
	}
	return nil
}
