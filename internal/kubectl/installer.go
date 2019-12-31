package kubectl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
	"github.com/dragonchain/dragonchain-installer/internal/downloader"
)

func kubectlIsInstalled() bool {
	return exec.Command("kubectl").Run() == nil
}

// InstallKubectlIfNecessary checks if kubectl is already installed, and installs it if necessary
func InstallKubectlIfNecessary() error {
	if kubectlIsInstalled() {
		fmt.Println("kubectl appears to already be installed")
		return nil
	}
	fmt.Println("kubectl is not installed. Installing now")
	if configuration.Windows {
		systemRoot, exists := os.LookupEnv("SYSTEMROOT")
		if !exists {
			return errors.New("Environment variable 'SYSTEMROOT' does not exist")
		}
		if err := downloader.DownloadFile(filepath.Join(systemRoot, "kubectl.exe"), configuration.WindowsKubectlLink); err != nil {
			return err
		}
	} else if configuration.Macos || configuration.Linux {
		tempDir, err := ioutil.TempDir("", "dcinstaller")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		tempPath := filepath.Join(tempDir, "kubectl")
		var allowExecute os.FileMode = 0775
		unixBinaryPath := filepath.Join("/", "usr", "local", "bin", "kubectl")
		downloadLink := configuration.LinuxKubectlLink
		if configuration.ARM64 {
			downloadLink = configuration.LinuxKubectlArm64Link
		} else if configuration.Macos {
			downloadLink = configuration.MacosKubectlLink
		}
		if err := downloader.DownloadFile(tempPath, downloadLink); err != nil {
			return err
		}
		if err := os.Chmod(tempPath, allowExecute); err != nil {
			return err
		}
		if configuration.Macos {
			// Move kubectl executable into /usr/local/bin (no elevated permissions required on macos)
			if err := os.Rename(tempPath, unixBinaryPath); err != nil {
				return err
			}
		} else {
			// Move kubectl executable into /usr/local/bin (require sudo on linux)
			cmd := exec.Command("sudo", "mv", tempPath, unixBinaryPath)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				return err
			}
		}
	} else {
		log.Fatal("Unsupported operating system")
	}
	// Should be installed at this point; if not, something is wrong
	if !kubectlIsInstalled() {
		return errors.New("Kubectl failed to install")
	}
	return nil
}
