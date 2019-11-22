package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/ini.v1"
)

func credentialFolderPath() (string, error) {
	if Windows {
		localAppData, exists := os.LookupEnv("LOCALAPPDATA")
		if !exists {
			return "", errors.New("Environment variable 'LOCALAPPDATA' does not exist")
		}
		return filepath.Join(localAppData, "dragonchain"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.New("Could not get home directory:\n" + err.Error())
	}
	return filepath.Join(home, ".dragonchain"), nil
}

func credentialFilePath() (string, error) {
	credentialFolder, err := credentialFolderPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(credentialFolder, "credentials"), nil
}

// Ensures that the file for dragonchain credentials exists
func ensureCredentialFile() error {
	folder, err := credentialFolderPath()
	if err != nil {
		return errors.New("Could not get credential folder:\n" + err.Error())
	}
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return errors.New("Failed to create dragonchain credentials configuration folder:\n" + err.Error())
	}
	filePath, err := credentialFilePath()
	if err != nil {
		return errors.New("Could not get credential file path:\n" + err.Error())
	}
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			// Create empty file since it does not exist
			_, err := os.Create(filePath)
			if err != nil {
				return errors.New("Failure to create file " + filePath + ":\n" + err.Error())
			}
		} else {
			return errors.New("Could not confirm existence of " + filePath + ":\n" + err.Error())
		}
	}
	return nil
}

// InstallDragonchainCredentials installs the credentials for this dragonchain to the local config to be used by sdk/cli tool, etc
func InstallDragonchainCredentials(config *Configuration, pubID string) error {
	fmt.Println("Installing new chain credentials for local use")
	// Make sure credentials file exists before reading it
	if err := ensureCredentialFile(); err != nil {
		return err
	}
	credentialsFile, err := credentialFilePath()
	if err != nil {
		return err
	}
	cfg, err := ini.Load(credentialsFile)
	if err != nil {
		return errors.New("Error loading credentials file:\n" + err.Error())
	}
	cfg.Section(pubID).Key("auth_key_id").SetValue(config.HmacID)
	cfg.Section(pubID).Key("auth_key").SetValue(config.HmacKey)
	/* Set the endpoint to the forwarded port from the VM on localhost

	In the future, we should consider allowing configuring this with the public endpoint, as using localhost won't easily support https,
	however using the public endpoint won't work for networks without some sort of NAT hairpinning/loopback support on the router */
	cfg.Section(pubID).Key("endpoint").SetValue("http://localhost:" + strconv.Itoa(config.Port))
	// Set default chain to this one
	if SetDefaultCredentials {
		cfg.Section("default").Key("dragonchain_id").SetValue(pubID)
	}
	if err := cfg.SaveTo(credentialsFile); err != nil {
		return errors.New("Error saving credentials file " + credentialsFile + ":\n" + err.Error())
	}
	return nil
}
