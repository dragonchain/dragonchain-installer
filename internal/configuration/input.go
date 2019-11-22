package configuration

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/dchest/uniuri"
)

// Configuration is all of the data needed to configure a new chain
type Configuration struct {
	Level             int
	Name              string
	EndpointURL       string
	Port              int
	InternalID        string
	RegistrationToken string
	PrivateKey        string
	HmacID            string
	HmacKey           string
}

var lowerCharNum = []byte("abcdefghijklmnopqrstuvxyz0123456789")

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
		if parsedPort < 30000 || parsedPort > 32767 {
			return -1, errors.New("Port must be between 30000 and 32767")
		}
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
	endpoint += ":" + strconv.Itoa(port)
	return endpoint, nil
}

// PromptForUserConfiguration get user input for all the necessary configurable variables of a Dragonchain
func PromptForUserConfiguration() (*Configuration, error) {
	// Get desired level
	level, err := getLevel()
	if err != nil {
		return nil, err
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
	// Construct and return the config object
	config := new(Configuration)
	config.Level = level
	config.Name = name
	config.EndpointURL = endpoint
	config.Port = port
	config.InternalID = internalID
	config.RegistrationToken = registrationToken
	return config, nil
}
