package kubernetes_infra

import (
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/exec/server"
	"k8s.io/client-go/tools/remotecommand"
)

type KubernetesExecSession struct {
	*server.ExecSessionBase

	Executor remotecommand.Executor
	SizeChan chan remotecommand.TerminalSize
}

func NewKubernetesExecSession(machineExec *model.MachineExec, executor remotecommand.Executor) *KubernetesExecSession {
	return &KubernetesExecSession{
		ExecSessionBase: server.NewExecSessionBase(machineExec),
		Executor: executor,
		SizeChan : make(chan remotecommand.TerminalSize),
	}
}

func (KubernetesExecSession) Stream() {
	// todo complete
}
