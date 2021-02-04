package configuration

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dchest/uniuri"
)

// Configuration is all of the data needed to configure a new chain
type Configuration struct {
	Level             int    `json:"Level"`
	Name              string `json:"Name"`
	EndpointURL       string `json:"EndpointURL"`
	Port              int    `json:"Port"`
	InternalID        string `json:"InternalID"`
	RegistrationToken string `json:"RegistrationToken"`
	UseVM             bool   `json:"UseVM"`
	InstallKubernetes bool	 `json:"InstallKubernetes"`
	Stage							string `json:"Stage"`
	PrivateKey        string
	HmacID            string
	HmacKey           string
}

var lowerCharNum = []byte("abcdefghijklmnopqrstuvxyz0123456789")

func configFilePath() (string, error) {
	credentialFolder, err := credentialFolderPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(credentialFolder, "installation_config"), nil
}

func checkExistingConfig() (*Configuration, error) {
	existingConfigFile, err := configFilePath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(existingConfigFile); os.IsNotExist(err) {
		return nil, err
	}
	file, err := ioutil.ReadFile(existingConfigFile)
	if err != nil {
		return nil, err
	}
	existingConf := new(Configuration)
	if err = json.Unmarshal([]byte(file), existingConf); err != nil {
		return nil, err
	}
	return existingConf, nil
}

func getPublicIP() (string, error) {
	resp, err := http.Get("https://ifconfig.co/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	bodyStr := string(body)
	return strings.TrimSuffix(bodyStr, "\n"), nil
}

func getUserInput(question string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(question)
	text, err := reader.ReadString('\n')
	if err != nil {
		return "", errors.New("Error reading input:\n" + err.Error())
	}
	text = strings.TrimSuffix(text, "\r\n") // Handle windows-style newlines
	text = strings.TrimSuffix(text, "\n")   // Handle unix-style newlines
	return text, nil
}

func getLevel() (int, error) {
	strLevel, err := getUserInput("What level chain would you like to create? [1-5]: ")
	if err != nil {
		return -1, err
	}
	level, err := strconv.ParseInt(strLevel, 10, 64)
	if err != nil {
		return -1, errors.New("Couldn't parse provided level into integer:\n" + err.Error())
	}
	if level < 1 || level > 5 {
		return -1, errors.New("Level must be between 1 and 5")
	}
	return int(level), nil
}

func getName() (string, error) {
	name, err := getUserInput("What name would you like for this chain? ")
	if err != nil {
		return "", err
	}
	nameRegex := `^[a-z][a-z0-9-_]{0,62}$`
	matched, err := regexp.MatchString(nameRegex, name)
	if err != nil {
		return "", errors.New("Failed to perform regex " + nameRegex + " on " + name)
	}
	if !matched {
		return "", errors.New("Provided name is not valid; Must match regex: " + nameRegex)
	}
	return name, nil
}

func getInternalID() (string, error) {
	internalID, err := getUserInput("Input the Chain ID for this chain (from Dragonchain console for Dragonnet support, otherwise leave empty): ")
	if err != nil {
		return "", err
	}
	if internalID == "" {
		// generate random arbitrary string if not provided
		internalID = uniuri.NewLenChars(16, lowerCharNum)
		fmt.Println("Defaulting to randomly generated: " + internalID)
	}
	return internalID, nil
}

func getRegistrationToken() (string, error) {
	registrationToken, err := getUserInput("Input the matchmaking token for this chain (from Dragonchain console for Dragonnet support, otherwise leave empty): ")
	if err != nil {
		return "", err
	}
	if registrationToken == "" {
		// generate random arbitrary string if not provided
		registrationToken = uniuri.NewLenChars(16, lowerCharNum)
		fmt.Println("Defaulting to randomly generated: " + registrationToken)
	}
	return registrationToken, nil
}

func getPort() (int, error) {
	portStr, err := getUserInput("What port would you like to run the dragonchain on? [30000-32767]: ")
	if err != nil {
		return -1, err
	}
	port := 30000 // default port
	if portStr == "" {
		fmt.Println("Defaulting port to " + strconv.Itoa(port))
	} else {
		parsedPort, err := strconv.ParseInt(portStr, 10, 64)
		if err != nil {
			return -1, errors.New("Couldn't parse provided port into integer:\n" + err.Error())
		}
		// if parsedPort < 30000 || parsedPort > 32767 {
		// 	return -1, errors.New("Port must be between 30000 and 32767")
		// }
		port = int(parsedPort)
	}
	return port, nil
}

func getEndpoint(port int) (string, error) {
	endpoint, err := getUserInput("What endpoint would you like to broadcast that this chain is available at? (i.e. http://my.domain) (Leave blank to find your public ip and use that): ")
	if err != nil {
		return "", err
	}
	if endpoint == "" {
		// Default endpoint to auto-retrieved public ip if not provided
		pubIP, err := getPublicIP()
		if err != nil {
			return "", errors.New("Issue getting public IP:\n" + err.Error())
		}
		fmt.Println("Defaulting to endpoint with public ip " + pubIP)
		endpoint = "http://" + pubIP
	} else {
		validEndpointRegex := `^http(s)?://(((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))|((([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])))$`
		matched, err := regexp.MatchString(validEndpointRegex, endpoint)
		if err != nil {
			return "", errors.New("Failed to perform regex " + validEndpointRegex + " on " + endpoint)
		}
		if !matched {
			return "", errors.New("Provided endpoint is not valid; Must look something like: http://a.b (dns name or ip are valid)")
		}
	}
	// add selected port to the endpoint
	if port != 80 && port != 8080 {
		endpoint += ":" + strconv.Itoa(port)
	}
	return endpoint, nil
}

func getVMDriver() (bool, error) {
	if !Linux {
		// VM Driver must be used if not on linux
		return true, nil
	}
	if !AMD64 {
		// VM Driver false is required if not AMD64
		return false, nil
	}
	driver, err := getUserInput("Would you like to use your machine's native docker and run kubernetes outside of a VM? (yes/no) ")
	if err != nil {
		return false, err
	}
	driver = strings.ToLower(driver)
	if driver == "y" || driver == "yes" {
		// ensure docker is installed and running
		cmd := exec.Command("sudo", "docker", "version")
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			return false, errors.New("Error checking for running docker daemon:\n" + err.Error())
		}
		return false, nil
	} else if driver == "n" || driver == "no" {
		return true, nil
	}
	return false, errors.New("Must answer yes/no")
}

func shouldInstallKubernetes() (bool, error) {
	answer, err := getUserInput("Would you like to install on existing kubernetes architecture? (yes/no) ")
	if err != nil {
		return false, err
	}
	answer = strings.ToLower(answer)
	if answer == "y" || answer == "yes" {
		return true, nil
	} else if answer == "n" || answer == "no" {
		return false, nil
	}
	return false, errors.New("Must answer yes/no")
}

func getStage() (string) {
	stage := "dev"
	answer, _ := getUserInput("Which stage would you like to use for your chain? (prod/dev) ")
	answer = strings.ToLower(answer)
	if answer == "p" || answer == "prod" {
		stage = "prod"
	}
	return stage
}

// PromptForUserConfiguration get user input for all the necessary configurable variables of a Dragonchain
func PromptForUserConfiguration() (*Configuration, error) {
	// Check for existing configuration from previous run first
	existingConf, err := checkExistingConfig()
	if err == nil {
		answer, err := getUserInput(`Existing config found:
			Level: ` + strconv.Itoa(existingConf.Level) + `
			Name: ` + existingConf.Name + `
			EndpointURL: ` + existingConf.EndpointURL + `
			Port: ` + strconv.Itoa(existingConf.Port) + `
			ChainID: ` + existingConf.InternalID + `
			MatchmakingToken: ` + existingConf.RegistrationToken + `
			UseVM: ` + strconv.FormatBool(existingConf.UseVM) + `
			InstallKubernetes: ` + strconv.FormatBool(existingConf.InstallKubernetes) + `
			Stage: ` + existingConf.Stage + `
			Would you like to use this config? (yes/no) `)
		if err != nil {
			return nil, err
		}
		answer = strings.ToLower(answer)
		if answer == "y" || answer == "yes" {
			return existingConf, nil
		} else if answer == "n" || answer == "no" {
			// Nothing happens, simply continue as normal
		} else {
			return nil, errors.New("Must answer yes/no")
		}
	}
	stage := getStage()
	installKubernetes, err := shouldInstallKubernetes()
	if err != nil {
		return nil, err
	}
	// Get desired vm usage
	vmDriver, err := getVMDriver()
	if err != nil {
		return nil, err
	}
	// Get desired level
	level, err := getLevel()
	if err != nil {
		return nil, err
	}
	if level == 1 && !AMD64 {
		return nil, errors.New("Level 1 chains are not supported on your cpu architecture")
	}
	// Get desired name
	name, err := getName()
	if err != nil {
		return nil, err
	}
	// Get internal id
	internalID, err := getInternalID()
	if err != nil {
		return nil, err
	}
	// Get registration token
	registrationToken, err := getRegistrationToken()
	if err != nil {
		return nil, err
	}
	// Get the desired port for the dragonchain
	port, err := getPort()
	if err != nil {
		return nil, err
	}
	// Get the desired endpoint
	endpoint, err := getEndpoint(port)
	if err != nil {
		return nil, err
	}
	// Construct and save the config object
	config := new(Configuration)
	config.Level = level
	config.Name = name
	config.EndpointURL = endpoint
	config.Port = port
	config.InternalID = internalID
	config.RegistrationToken = registrationToken
	config.UseVM = vmDriver
	config.InstallKubernetes = installKubernetes
	config.Stage = stage
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	folder, err := credentialFolderPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return nil, err
	}
	configFile, err := configFilePath()
	if err != nil {
		return nil, err
	}
	if err = ioutil.WriteFile(configFile, configJSON, 0664); err != nil {
		return nil, err
	}
	return config, nil
}
