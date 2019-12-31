package helm

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

func helmIsInstalled() bool {
	return exec.Command("helm").Run() == nil
}

// InstallHelmIfNecessary checks if helm is already installed, and installs it if necessary
func InstallHelmIfNecessary() error {
	if helmIsInstalled() {
		fmt.Println("helm appears to already be installed")
		return nil
	}
	fmt.Println("helm is not installed. Installing now")
	tempDir, err := ioutil.TempDir("", "dcinstaller")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)
	if configuration.Windows {
		systemRoot, exists := os.LookupEnv("SYSTEMROOT")
		if !exists {
			return errors.New("Environment variable 'SYSTEMROOT' does not exist")
		}
		helmZip := filepath.Join(tempDir, "helm.zip")
		if err := downloader.DownloadFile(helmZip, configuration.WindowsHelmLink); err != nil {
			return err
		}
		// Extract the zip file
		cmd := exec.Command("powershell", "-nologo", "-noprofile", "-command", "& { Add-Type -A 'System.IO.Compression.FileSystem'; [IO.Compression.ZipFile]::ExtractToDirectory('"+helmZip+"', '"+tempDir+"'); }")
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		// Move the helm file into the system path to install it
		if err := os.Rename(filepath.Join(tempDir, "windows-amd64", "helm.exe"), filepath.Join(systemRoot, "helm.exe")); err != nil {
			return err
		}
	} else if configuration.Macos || configuration.Linux {
		tempZip := filepath.Join(tempDir, "helm.tar.gz")
		downloadLink := configuration.LinuxHelmLink
		extractedFolder := "linux-amd64"
		if configuration.ARM64 {
			downloadLink = configuration.LinuxHelmArm64Link
			extractedFolder = "linux-arm64"
		} else if configuration.Macos {
			downloadLink = configuration.MacosHelmLink
			extractedFolder = "darwin-amd64"
		}
		unixInstallPath := filepath.Join("/", "usr", "local", "bin", "helm")
		// Download the helm gzip package
		if err := downloader.DownloadFile(tempZip, downloadLink); err != nil {
			return err
		}
		// Extract the package
		cmd := exec.Command("tar", "-xzf", tempZip, "-C", tempDir)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
		if configuration.Macos {
			// Move helm executable into /usr/local/bin (no elevated permissions required on macos)
			if err := os.Rename(filepath.Join(tempDir, extractedFolder, "helm"), unixInstallPath); err != nil {
				return err
			}
		} else {
			// Move helm executable into /usr/local/bin (require sudo on linux)
			cmd := exec.Command("sudo", "mv", filepath.Join(tempDir, extractedFolder, "helm"), unixInstallPath)
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
	if !helmIsInstalled() {
		return errors.New("Helm failed to install")
	}
	return nil
}
