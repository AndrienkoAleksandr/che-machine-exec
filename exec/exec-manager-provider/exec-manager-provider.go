package exec_manager_provider

import (
	"github.com/eclipse/che-machine-exec/api/model"
	"os"
)

// infra enums
type INFRA string
const (
	DOCKER INFRA = "docker"
	KUBERNETES INFRA = "kubernetes"
)

type ExecManager interface {
	getInfra() INFRA

	Create() (int, error)
	Check() (int, error)
	Attach() (*model.MachineExec, error)
	Resize() error
}

func CreateExecManager() (ExecManager, error)  {
	if IsKubernetesInfra() {
		//return kubernetes machine exec impl
	}
	if IsDockerInfra() {
		// return docker machine exec impl
	}
	// todo cache exec manager

	return nil, nil
}

func IsKubernetesInfra() bool {
	stat, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if err == nil && stat.IsDir() {
		return true
	}

	return false
}

func IsDockerInfra() bool {
	stat, err := os.Stat("/var/run/docker.sock")
	if err == nil && stat.Mode().IsRegular() {
		return true
	}

	return false
}
