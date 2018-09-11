package kubernetes_infra

import (
	"fmt"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"log"
)

// Kubernetes pty handler
type KubernetesPtyHandler struct {
	*model.InOutHandlerBase

	exec *model.MachineExec
	//hide it in the package scope!!!!
	sizeChan chan remotecommand.TerminalSize
	executor remotecommand.Executor
}

func NewPtyHandler(exec *model.MachineExec, executor remotecommand.Executor) *KubernetesPtyHandler {
	sizeChan := make(chan remotecommand.TerminalSize)

	msgChan := make(chan []byte)
	inOutHandler := &model.InOutHandlerBase{MsgChan:msgChan}

	return &KubernetesPtyHandler{exec: exec, sizeChan:sizeChan, executor:executor, InOutHandlerBase:inOutHandler}
}

func (ptyH KubernetesPtyHandler) execIsAttached() bool {
	panic("implement me")
}

func (ptyH KubernetesPtyHandler) Stream() {
	panic("implement me")
}

func (ptyH KubernetesPtyHandler) Read(p []byte) (int, error) {
	data := <-ptyH.MsgChan

	return copy(p, data), nil
}

func (ptyH KubernetesPtyHandler) Write(p []byte) (int, error) {
	fmt.Println(" Output! " + string(p))

	ptyH.exec.WriteDataToWsConnections(p)

	return len(p), nil
}

func (ptyH KubernetesPtyHandler) Next() *remotecommand.TerminalSize {
	log.Println("next size")
	select {
	case size := <-ptyH.sizeChan:
		log.Println("So next size will be ", size)
		return &size
	}
}
