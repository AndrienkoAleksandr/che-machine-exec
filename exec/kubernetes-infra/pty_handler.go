package kubernetes_infra

import (
	"fmt"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"log"
)

// Kubernetes pty handler
type KubernetesPtyHandlerImpl struct {
	machineExec *model.MachineExec
	//hide it in the package scope!!!!
	SizeChan chan remotecommand.TerminalSize
	Executor remotecommand.Executor
}

func NewPtyHandler(machineExec *model.MachineExec, executor remotecommand.Executor) *KubernetesPtyHandlerImpl {
	sizeChan := make(chan remotecommand.TerminalSize)
	return &KubernetesPtyHandlerImpl{machineExec: machineExec, SizeChan:sizeChan, Executor:executor}
}

func (t KubernetesPtyHandlerImpl) Read(p []byte) (int, error) {
	data := <-t.machineExec.MsgChan

	return copy(p, data), nil
}

func (t KubernetesPtyHandlerImpl) Write(p []byte) (int, error) {
	fmt.Println(" Output! " + string(p))

	t.machineExec.WriteDataToWsConnections(p)

	return len(p), nil
}

func (t KubernetesPtyHandlerImpl) Next() *remotecommand.TerminalSize {
	log.Println("next size")
	select {
	case size := <-t.SizeChan:
		log.Println("So next size will be ", size)
		return &size
	}
}
