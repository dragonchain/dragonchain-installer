package virtualbox

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
	"github.com/dragonchain/dragonchain-installer/internal/downloader"
)

func virtualBoxIsInstalled() bool {
	return exec.Command(vboxManageExecutable(), "--version").Run() == nil
}

// InstallVirtualBoxIfNecessary checks if virtualbox is already installed, and installs it if necessary
func InstallVirtualBoxIfNecessary() error {
	if !configuration.AMD64 {
		return errors.New("Cannot install virtualbox on non-amd64 architecture")
	}
	if virtualBoxIsInstalled() {
		fmt.Println("virtualbox appears to already be installed")
		return nil
	}
	fmt.Println("virtualbox is not installed. Installing now")
	if configuration.Windows {
		if err := installVirtualBoxWindows(); err != nil {
			return err
		}
	} else if configuration.Macos {
		if err := installVirtualBoxMacos(); err != nil {
			return err
		}
	} else if configuration.Linux {
		if err := installVirtualBoxLinux(); err != nil {
			return err
		}
	} else {
		log.Fatal("Unsupported operating system")
	}
	// Should be installed at this point; if not, something is wrong
	if !virtualBoxIsInstalled() {
		return errors.New("Virtualbox failed to install")
	}
	return nil
}

func installVirtualBoxLinux() error {
	// Create the temp dir for the download
	tempDir, err := ioutil.TempDir("", "dcinstaller")
	if err != nil {
		return errors.New("Creating temporary directory failed:\n" + err.Error())
	}
	defer os.RemoveAll(tempDir)
	// Download virtualbox
	fmt.Println("Downloading virtualbox")
	installerFile := filepath.Join(tempDir, "virtualbox.run")
	if err := downloader.DownloadFile(installerFile, configuration.LinuxVirtualboxLink); err != nil {
		return errors.New("Downloading virtualbox failed:\n" + err.Error())
	}
	// Set execution permissions
	var allowExecute os.FileMode = 0775
	if err := os.Chmod(installerFile, allowExecute); err != nil {
		return errors.New("Setting execute permission on downloaded file failed:\n" + err.Error())
	}
	// Run the installer (require sudo)
	fmt.Println("Installing Virtualbox")
	cmd := exec.Command("sudo", installerFile)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.New("Installing virtualbox failed:\n" + err.Error())
	}
	return nil
}

func installVirtualBoxMacos() error {
	// Create the temp dir for the download
	tempDir, err := ioutil.TempDir("", "dcinstaller")
	if err != nil {
		return errors.New("Creating temporary directory failed:\n" + err.Error())
	}
	defer os.RemoveAll(tempDir)
	// Download virtualbox
	fmt.Println("Downloading virtualbox")
	installerFile := filepath.Join(tempDir, "virtualbox.dmg")
	if err := downloader.DownloadFile(installerFile, configuration.MacosVirtualboxLink); err != nil {
		return errors.New("Downloading virtualbox failed:\n" + err.Error())
	}
	// Mount the dmg
	fmt.Println("Installing Virtualbox")
	cmd := exec.Command("hdiutil", "attach", installerFile)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Mounting virtualbox dmg failed:\n" + err.Error())
	}
	defer exec.Command("hdiutil", "detach", "/Volumes/VirtualBox").Run()
	// Copy the pkg and remove its extended attributes for installation
	virtualBoxPkg := filepath.Join(tempDir, "virtualbox.pkg")
	cmd = exec.Command("cp", "-f", "/Volumes/VirtualBox/VirtualBox.pkg", virtualBoxPkg)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error copying virtualbox pkg file:\n" + err.Error())
	}
	cmd = exec.Command("xattr", "-c", virtualBoxPkg)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error removing extended attributes from pkg:\n" + err.Error())
	}
	// Install pkg (with sudo, prompting for password if necessary)
	cmd = exec.Command("sudo", "installer", "-package", virtualBoxPkg, "-target", "/")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return errors.New("Installing virtualbox pkg failed (are you root?):\n" + err.Error())
	}
	return nil
}

func installVirtualBoxWindows() error {
	tempRoot, exists := os.LookupEnv("TEMP")
	if !exists {
		return errors.New("Environment variable 'TEMP' could not be found")
	}
	vboxTemp := filepath.Join(tempRoot, "VirtualBox")
	// Create the temp dir for the download
	tempDir, err := ioutil.TempDir("", "dcinstaller")
	if err != nil {
		return errors.New("Creating temporary directory failed:\n" + err.Error())
	}
	defer os.RemoveAll(tempDir)
	// Download the installer
	fmt.Println("Downloading virtualbox")
	exeFile := filepath.Join(tempDir, "virtualbox.exe")
	if err := downloader.DownloadFile(exeFile, configuration.WindowsVirtualboxLink); err != nil {
		return errors.New("Downloading virtualbox failed:\n" + err.Error())
	}
	// Extract the msi installer
	fmt.Println("Installing Virtualbox")
	cmd := exec.Command(exeFile, "-extract", "-silent")
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Extracting msi installer from virtualbox exe failed:\n" + err.Error())
	}
	defer os.RemoveAll(vboxTemp)
	// Find the correct extracted msi
	files, err := ioutil.ReadDir(vboxTemp)
	if err != nil {
		return errors.New("Reading from virtualbox temp directory failed:\n" + err.Error())
	}
	msiToUse := ""
	for _, f := range files {
		if name := f.Name(); strings.HasSuffix(name, "amd64.msi") {
			msiToUse = name
		}
	}
	if msiToUse == "" {
		return errors.New("Couldn't find extracted msi to install virtualbox")
	}
	// Install the extracted msi
	cmd = exec.Command("msiexec", "/i", filepath.Join(vboxTemp, msiToUse), "/quiet", "/qn", "/norestart", "/log", filepath.Join(tempDir, "vbox_install.log"))
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Running msiexec on extracted virtualbox installer failed (are you running as administrator?):\n" + err.Error())
	}
	return nil
}
