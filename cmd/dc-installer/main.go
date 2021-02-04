package main

import (
	"fmt"
	"github.com/dragonchain/dragonchain-installer/internal/kubectl"
	"os"
	"strings"

	"github.com/dragonchain/dragonchain-installer/internal/configuration"
	"github.com/dragonchain/dragonchain-installer/internal/dragonchain"
	"github.com/dragonchain/dragonchain-installer/internal/dragonnet"
	"github.com/dragonchain/dragonchain-installer/internal/helm"
	"github.com/dragonchain/dragonchain-installer/internal/minikube"
	"github.com/dragonchain/dragonchain-installer/internal/upnp"
	"github.com/dragonchain/dragonchain-installer/internal/virtualbox"
)

func fatalLog(v ...interface{}) {
	fmt.Println(v...)
	if configuration.Windows {
		// If windows, require pressing enter before exiting
		fmt.Print("\nFinished. Press enter to exit program\n")
		fmt.Scanln()
	}
	os.Exit(1)
}

func installer() {
	fmt.Print("Starting dragonchain installer\n")
	config, err := configuration.PromptForUserConfiguration()
	if err != nil {
		fatalLog(err)
	}
	fmt.Print("Starting dragonchain installer\nChecking for required dependencies\n\n")
	if config.InstallKubernetes {
		if err := kubectl.InstallKubectlIfNecessary(); err != nil {
			fatalLog(err)
		}
		if err := helm.InstallHelmIfNecessary(); err != nil {
			fatalLog(err)
		}
		if err := minikube.InstallMinikubeIfNecessary(); err != nil {
			fatalLog(err)
		}
		fmt.Print("\nBase dependencies installed\nConfiguring dependencies now\n\n")
		if config.UseVM {
			fmt.Print("Virtualbox required for minikube VM. Checking and installing if necessary\n")
			if err := virtualbox.InstallVirtualBoxIfNecessary(); err != nil {
				fatalLog(err)
			}
		}
		if err := minikube.StartMinikubeCluster(config.UseVM); err != nil {
			fatalLog(err)
		}
	}
	if err := helm.InitializeHelm(); err != nil {
		fatalLog(err)
	}
	if err := dragonchain.SetupDragonchainPreReqs(config); err != nil {
		fatalLog(err)
	}
	if config.UseVM && config.InstallKubernetes {
		if err := virtualbox.ConfigureVirtualboxVM(config); err != nil {
			fatalLog(err)
		}
	}
	fmt.Print("\nConfiguration of dependencies complete\nNow installing Dragonchain\n")
	if err := dragonchain.InstallDragonchain(config); err != nil {
		fatalLog(err)
	}
	fmt.Print("Installation Complete\n\nGetting public ID\n")
	pubID, err := dragonchain.GetDragonchainPublicID(config)
	if err != nil {
		fatalLog(err)
	}
	fmt.Print("Dragonchain public id is: " + pubID + "\n\n")
	if config.InstallKubernetes {
		startCommand, stopCommand := minikube.FriendlyStartStopCommand(config.UseVM)
		fmt.Print("In order to stop the dragonchain, run the following command in a terminal:\n" + stopCommand + "\n\n")
		fmt.Print("In order to restart the dragonchain, run the following command in a terminal:\n" + startCommand + "\n\n")
	}
	if err := configuration.InstallDragonchainCredentials(config, pubID); err != nil {
		fatalLog(err)
	}
	fmt.Print("Checking dragon net for proper chain configuration\n")
	if err := dragonnet.CheckDragonNetConfiguration(pubID); err != nil {
		if strings.HasPrefix(err.Error(), "Although registered") {
			// If only issue with registration is that chain is registered, but not reachable (potential port-forward issue), try upnp
			fmt.Print("Chain is registered, but does not seem reachable. Trying to automatically port-forward with upnp\n")
			if upnpErr := upnp.AddUPNPPortMapping(config.Port); upnpErr != nil {
				fmt.Print("Could not port forward with upnp:\n" + upnpErr.Error())
			} else {
				fmt.Print("Port forward with upnp successful, checking dragonnet registration again\n")
				err = dragonnet.CheckDragonNetConfiguration(pubID)
			}
		}
		if err != nil {
			fatalLog("\nDragonchain is installed and may be working locally, but dragon net configuration seems invalid\n", err)
		}
	}
	// Successful installation and dragon net configuration
	fmt.Print("\nChain is installed, running, and operating correctly with Dragon Net!\n")
	if configuration.Windows {
		// If windows, require pressing enter before exiting
		fmt.Print("\nFinished. Press enter to exit program\n")
		fmt.Scanln()
	}
}

func main() {
	if !(configuration.Windows || configuration.Linux || configuration.Macos) {
		fatalLog("Unsupported OS")
	}
	if !configuration.AMD64 {
		// Only non-amd64 architecture supported is arm64 on linux
		if !(configuration.ARM64 && configuration.Linux) {
			fatalLog("Unsupported OS/Architecture")
		}
		fmt.Println("WARNING!!! ARM64 support is currently experimental and not fully working/supported.")
	}
	if len(os.Args) > 1 && (os.Args[1] == "-V" || os.Args[1] == "--version" || os.Args[1] == "version") {
		fmt.Println(configuration.Version)
	} else {
		// Don't allow the program to run as root
		if os.Geteuid() == 0 {
			fatalLog("Do not run this program as root. Run it as your regular user")
		}
		installer()
	}
	os.Exit(0)
}
