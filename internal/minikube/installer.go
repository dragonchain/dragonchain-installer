package minikube

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

func minikubeIsInstalled() bool {
	return exec.Command("minikube").Run() == nil
}

// InstallMinikubeIfNecessary checks if minikube is already installed, and installs it if necessary
func InstallMinikubeIfNecessary() error {
	if minikubeIsInstalled() {
		fmt.Println("minikube appears to already be installed")
		return nil
	}
	fmt.Println("minikube is not installed. Installing now")
	if configuration.Windows {
		systemRoot, exists := os.LookupEnv("SYSTEMROOT")
		if !exists {
			return errors.New("Environment variable 'SYSTEMROOT' does not exist")
		}
		if err := downloader.DownloadFile(filepath.Join(systemRoot, "minikube.exe"), configuration.WindowsMinikubeLink); err != nil {
			return err
		}
	} else if configuration.Macos || configuration.Linux {
		tempDir, err := ioutil.TempDir("", "dcinstaller")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tempDir)
		tempPath := filepath.Join(tempDir, "minikube")
		var allowExecute os.FileMode = 0775
		unixBinaryPath := filepath.Join("/", "usr", "local", "bin", "minikube")
		downloadLink := configuration.LinuxMinikubeLink
		if configuration.ARM64 {
			downloadLink = configuration.LinuxMinikubeArm64Link
		} else if configuration.Macos {
			downloadLink = configuration.MacosMinikubeLink
		}
		if err := downloader.DownloadFile(tempPath, downloadLink); err != nil {
			return err
		}
		if err := os.Chmod(tempPath, allowExecute); err != nil {
			return err
		}
		if configuration.Macos {
			// Move minikube executable into /usr/local/bin (no elevated permissions required on macos)
			if err := os.Rename(tempPath, unixBinaryPath); err != nil {
				return err
			}
		} else {
			// Move minikube executable into /usr/local/bin (require sudo on linux)
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
	if !minikubeIsInstalled() {
		return errors.New("Minikube failed to install")
	}
	return nil
}
