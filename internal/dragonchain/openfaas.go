package dragonchain

import (
	"bytes"
	"errors"
	"os"
	"os/exec"

	"github.com/dchest/uniuri"
	"github.com/dragonchain/dragonchain-installer/internal/configuration"
)

var openfaasNamespacesYaml = []byte(`
apiVersion: v1
kind: Namespace
metadata:
  name: openfaas
  labels:
    role: openfaas-system
    access: openfaas-system
    istio-injection: enabled
---
apiVersion: v1
kind: Namespace
metadata:
  name: openfaas-fn
  labels:
    istio-injection: enabled
    role: openfaas-fn`)

var serviceAccountYaml = []byte(`
apiVersion: v1
kind: ServiceAccount
metadata:
  name: openfaas-builder
  namespace: dragonchain
automountServiceAccountToken: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: dragonchain
  name: openfaas-builder
rules:
- apiGroups: ["batch"]
  resources: ["jobs", "jobs/status"]
  verbs: ["create", "get", "delete"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  namespace: dragonchain
  name: openfaas-builder
subjects:
- kind: ServiceAccount
  name: openfaas-builder
  apiGroup: ""
roleRef:
  kind: Role
  name: openfaas-builder
  apiGroup: rbac.authorization.k8s.io`)

func openfaasServiceAccountExists() bool {
	return exec.Command("kubectl", "get", "serviceaccount", "-n", "dragonchain", "openfaas-builder", "--context="+configuration.MinikubeContext).Run() == nil
}

func createOpenFaasDeployment() error {
	// Create the necessary namespaces
	cmd := exec.Command("kubectl", "apply", "--context="+configuration.MinikubeContext, "-f", "-")
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewBuffer(openfaasNamespacesYaml)
	if err := cmd.Run(); err != nil {
		return errors.New("Error creating openfaas namespaces:\n" + err.Error())
	}
	// Create the basic auth secrets
	secret := uniuri.NewLen(40)
	cmd = exec.Command("kubectl", "create", "secret", "generic", "basic-auth", "--from-literal=basic-auth-user=admin", "--from-literal=basic-auth-password="+secret, "-n", "openfaas", "--context="+configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error creating openfaas kubernetes secret:\n" + err.Error())
	}
	cmd = exec.Command("kubectl", "create", "secret", "generic", "openfaas-auth", "--from-literal=user=admin", "--from-literal=password="+secret, "-n", "dragonchain", "--context="+configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error creating openfaas kubernetes secret:\n" + err.Error())
	}
	// Install openfaas
	cmd = exec.Command("helm", "upgrade", "--install", "openfaas", "openfaas/openfaas", "--namespace", "openfaas", "--set", "basic_auth=true,generateBasicAuth=false,functionNamespace=openfaas-fn,async=false,exposeServices=false,alertmanager.create=false,prometheus.create=false", "--version", configuration.OpenfaasHelmVersion, "--kube-context", configuration.MinikubeContext)
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.New("Error helm deploying openfaas:\n" + err.Error())
	}
	return nil
}

func createOpenfaasBuilderServiceAccount() error {
	// Add the service account
	cmd := exec.Command("kubectl", "apply", "--context="+configuration.MinikubeContext, "-f", "-")
	cmd.Stderr = os.Stderr
	cmd.Stdin = bytes.NewBuffer(serviceAccountYaml)
	if err := cmd.Run(); err != nil {
		return errors.New("Error helm deploying openfaas:\n" + err.Error())
	}
	return nil
}
