package exec

import (
	"fmt"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/exec/docker-infra"
	"github.com/eclipse/che-machine-exec/exec/kubernetes-infra"
	"os"
)

// infra enums
//type INFRA string
//const (
//	DOCKER INFRA = "docker"
//	KUBERNETES INFRA = "kubernetes"
//)

var execManager ExecManager

type ExecManager interface {
	//getInfra() INFRA

	Create(*model.MachineExec) (int, error)
	Check(id int) (int, error)
	Attach(id int, conn *websocket.Conn) error
	Resize(id int, cols uint, rows uint) error
}

func CreateExecManager() ExecManager {
	var manager ExecManager

	if IsKubernetesInfra() {
		fmt.Println("Use kubernetes implementation")
		manager = kubernetes_infra.New()
	} else if IsDockerInfra() {
		fmt.Println("Use docker implementation")
		manager = docker_infra.New()
	}

	// todo what we should do in the case, when we have no implementation. Should we return stub, or only log error

	return manager
}

func GetExecManager() ExecManager {
	if execManager == nil {
		execManager = CreateExecManager()
	}
	return execManager
}

//todo rework this method should be hidden, starts with lower character
func IsKubernetesInfra() bool {
	stat, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount")
	if err == nil && stat.IsDir() {
		return true
	}

	return false
}

//todo rework this method should be hidden, starts with lower character
func IsDockerInfra() bool {
	stat, err := os.Stat("/var/run/docker.sock")
	if err == nil && !stat.Mode().IsRegular() && !stat.IsDir() {
		return true
	}

	return false
}
