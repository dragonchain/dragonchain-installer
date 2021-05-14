package dragonchain

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/dchest/uniuri"
	"github.com/dragonchain/dragonchain-installer/internal/configuration"
	"github.com/vsergeev/btckeygenie/btckey"
)

var allChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var upperChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

type kubectlSecretJSON struct {
	Data (struct {
		SecretString string `json:"SecretString"`
	}) `json:"data"`
}

type dcSecretJSON struct {
	PrivateKey string `json:"private-key"`
	HmacID     string `json:"hmac-id"`
	HmacKey    string `json:"hmac-key"`
}

func genRandomSecp256k1Key() (string, error) {
	priv, err := btckey.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(priv.ToBytes()), nil
}

func generateDragonchainSecrets() (string, string, string, error) {
	key, err := genRandomSecp256k1Key()
	if err != nil {
		return "", "", "", errors.New("Error generating new private key:\n" + err.Error())
	}
	hmacID := uniuri.NewLenChars(12, upperChars)
	hmacKey := uniuri.NewLenChars(43, allChars)
	fmt.Println("Root HMAC key details: ID: " + hmacID + " | KEY: " + hmacKey)
	return key, hmacID, hmacKey, nil
}

func dragonchainSecretName(internalID string) string {
	return "d-" + internalID + "-secrets"
}

func createDragonchainSecret(config *configuration.Configuration) error {
	key, hmacID, hmacKey, err := generateDragonchainSecrets()
	if err != nil {
		return err
	}
	// Set secrets on configuration struct
	config.PrivateKey = key
	config.HmacID = hmacID
	config.HmacKey = hmacKey
	secretJSON := "{\"private-key\":\"" + key + "\",\"hmac-id\":\"" + hmacID + "\",\"hmac-key\":\"" + hmacKey + "\",\"registry-password\":\"\"}"
	cmd := exec.Command("kubectl", "create", "secret", "generic", dragonchainSecretName(config.InternalID), "--from-literal=SecretString="+secretJSON, "-n", "dragonchain", "--context="+configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error adding secret for new dragonchain:\n" + err.Error())
	}
	return nil
}

func chainSecretExists(internalID string) bool {
	return exec.Command("kubectl", "get", "secret", "-n", "dragonchain", dragonchainSecretName(internalID), "--context="+configuration.MinikubeContext).Run() == nil
}

func getExistingSecret(config *configuration.Configuration) error {
	cmd := exec.Command("kubectl", "get", "secret", "-n", "dragonchain", dragonchainSecretName(config.InternalID), "-o", "json", "--context="+configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return errors.New("Error retrieving existing dragonchain secret:\n" + err.Error())
	}
	var kubeSecret kubectlSecretJSON
	if err := json.Unmarshal(output, &kubeSecret); err != nil {
		return errors.New("Error parsing dragonchain secret:\n" + err.Error())
	}
	decoded, err := base64.StdEncoding.DecodeString(kubeSecret.Data.SecretString)
	if err != nil {
		return errors.New("Error decoding base64 secret value:\n" + err.Error())
	}
	var dcSecrets dcSecretJSON
	if err := json.Unmarshal(decoded, &dcSecrets); err != nil {
		return errors.New("Error parsing dragonchain secret:\n" + err.Error())
	}
	config.PrivateKey = dcSecrets.PrivateKey
	config.HmacKey = dcSecrets.HmacKey
	config.HmacID = dcSecrets.HmacID
	return nil
}

// configuration.DragonchainHelmVersion

func listDirectories() {
	cmd := exec.Command("ls", "-la")
	output, _ := cmd.Output()
	fmt.Print(string(output[:]))
}

func upsertDragonchainHelmDeployment(config *configuration.Configuration) error {
	setStringStr := "global.environment.LEVEL=" + strconv.Itoa(config.Level)
	if config.Stage == "dev" {
		setStringStr = setStringStr + ",global.environment.STAGE=" + config.Stage
	}
	if config.S3Bucket != "" {
		setStringStr = setStringStr + ",global.environment.STORAGE_LOCATION=" + config.S3Bucket + ",global.environment.STORAGE_TYPE=s3"
	}
	setStr := "ingressEndpoint=eks.dragonchain.com,dragonchain.storage.spec.storageClassName=gp2,redis.storage.spec.storageClassName=gp2,redisearch.storage.spec.storageClassName=gp2,global.environment.DRAGONCHAIN_NAME=" + config.Name + ",global.environment.REGISTRATION_TOKEN=" + config.RegistrationToken + ",global.environment.INTERNAL_ID=" + config.InternalID + ",global.environment.DRAGONCHAIN_ENDPOINT=" + config.EndpointURL + ",service.port=" + strconv.Itoa(config.Port)
	if config.Level == 1 {
		setStr += ",faas.gateway=http://gateway.openfaas:8080,faas.mountFaasSecret=true,faas.registry=" + configuration.RegistryIP + ":" + strconv.Itoa(configuration.RegistryPort)
	}
	cmd := exec.Command("helm", "upgrade", "--install", "d-"+config.InternalID, "./dc-k8s-helm", "--namespace", "dragonchain", "--set-string", setStringStr, "--set", setStr, "--version", "1.0.9", "--kube-context", configuration.MinikubeContext)
	if !config.InstallKubernetes {
		cmd = exec.Command("helm", "upgrade", "--install", "-f", "./dragonchain-eks/dc-k8s-helm/values.eks.yaml", "d-"+config.InternalID, "./dragonchain-eks/dc-k8s-helm", "--namespace", "dragonchain", "--set-string", setStringStr, "--set", setStr, "--version", "1.0.9", "--kube-context", configuration.MinikubeContext)
	}
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error installing dragonchain helm chart:\n" + err.Error())
	}
	return nil
}

// InstallDragonchain installs the kubernetes resources for the dragonchain (and upgrades if it already exists)
func InstallDragonchain(config *configuration.Configuration) error {
	// Ensure kubernetes secret exists for this dragonchain
	if chainSecretExists(config.InternalID) {
		fmt.Println("Existing dragonchain secret for this id already exists. Reusing")
		if err := getExistingSecret(config); err != nil {
			return err
		}
	} else {
		fmt.Println("Creating new secret for this dragonchain id")
		if err := createDragonchainSecret(config); err != nil {
			return err
		}
	}
	// Actually install (or upgrade) the chain
	if err := upsertDragonchainHelmDeployment(config); err != nil {
		return err
	}
	fmt.Println("Dragonchain helm deployment complete. Waiting for chain to be ready.")
	// Wait for the deployment to be ready before continuing
	err := waitForDragonchainToBeReady(config)
	fmt.Print("\n")
	if err != nil {
		return err
	}
	return nil
}
