package kubernetes_infra

import (
	"fmt"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"github.com/eclipse/che-machine-exec/api/websocket/ws-conn"
)

type KubernetesExecStreamHandler struct {
	*model.InOutHandlerBase

	sizeChan chan remotecommand.TerminalSize
	executor remotecommand.Executor
}

func NewPtyHandler(executor remotecommand.Executor) *KubernetesExecStreamHandler {
	sizeChan := make(chan remotecommand.TerminalSize)

	msgChan := make(chan []byte)
	connsHandler := ws_conn.New()
	inOutHandler := &model.InOutHandlerBase{MsgChan:msgChan, ConnsHandler: connsHandler}

	return &KubernetesExecStreamHandler{
		sizeChan:sizeChan,
		executor:executor,
		InOutHandlerBase:inOutHandler,
	}
}

func (strH KubernetesExecStreamHandler) Stream(tty bool) error {
	return strH.executor.Stream(remotecommand.StreamOptions{
		Stdin:             strH,
		Stdout:            strH,
		Stderr:            strH,
		TerminalSizeQueue: strH,
		Tty:               tty,
	})
}

func (strH KubernetesExecStreamHandler) Read(p []byte) (int, error) {
	data := <-strH.MsgChan

	return copy(p, data), nil
}

func (strH KubernetesExecStreamHandler) Write(p []byte) (int, error) {
	fmt.Println(" Output! " + string(p))

	strH.Buffer.Write(p)
	strH.ConnsHandler.WriteDataToWsConnections(p)

	return len(p), nil
}

func (strH KubernetesExecStreamHandler) Next() *remotecommand.TerminalSize {
	log.Println("next size")
	select {
	case size := <-strH.sizeChan:
		log.Println("So next size will be ", size)
		return &size
	}
}
