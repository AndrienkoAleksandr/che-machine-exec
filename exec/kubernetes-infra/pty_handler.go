package kubernetes_infra

import (
	"fmt"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"github.com/eclipse/che-machine-exec/api/websocket/ws-conn"
)

// Kubernetes pty handler
type KubernetesExecStreamHandler struct {
	*model.InOutHandlerBase

	exec *model.MachineExec
	//hide it in the package scope!!!!
	sizeChan chan remotecommand.TerminalSize
	executor remotecommand.Executor
}

func NewPtyHandler(exec *model.MachineExec, executor remotecommand.Executor) *KubernetesExecStreamHandler {
	sizeChan := make(chan remotecommand.TerminalSize)

	msgChan := make(chan []byte)
	connsHandler := ws_conn.New()
	inOutHandler := &model.InOutHandlerBase{MsgChan:msgChan, ConnsHandler: connsHandler}

	return &KubernetesExecStreamHandler{
		exec: exec, sizeChan:sizeChan,
		executor:executor,
		InOutHandlerBase:inOutHandler,
	}
}

func (ptyH KubernetesExecStreamHandler) execIsAttached() bool {
	panic("implement me")
}

// rename ptyH
func (ptyH KubernetesExecStreamHandler) Stream(tty bool) error {
	return ptyH.executor.Stream(remotecommand.StreamOptions{
		Stdin:             ptyH,
		Stdout:            ptyH,
		Stderr:            ptyH,
		TerminalSizeQueue: ptyH,
		Tty:               tty,
	})
}

func (ptyH KubernetesExecStreamHandler) Read(p []byte) (int, error) {
	data := <-ptyH.MsgChan

	return copy(p, data), nil
}

func (ptyH KubernetesExecStreamHandler) Write(p []byte) (int, error) {
	fmt.Println(" Output! " + string(p))

	ptyH.ConnsHandler.WriteDataToWsConnections(p)

	return len(p), nil
}

func (ptyH KubernetesExecStreamHandler) Next() *remotecommand.TerminalSize {
	log.Println("next size")
	select {
	case size := <-ptyH.sizeChan:
		log.Println("So next size will be ", size)
		return &size
	}
}
