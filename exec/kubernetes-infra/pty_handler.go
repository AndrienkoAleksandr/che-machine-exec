package kubernetes_infra

import (
	"fmt"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"log"
)

// Kubernetes pty handler
type PtyHandlerImpl struct {
	machineExec *model.MachineExec
}

func (t PtyHandlerImpl) Read(p []byte) (int, error) {
	data := <-t.machineExec.MsgChan
	return copy(p, data), nil
}

func (t PtyHandlerImpl) Write(p []byte) (int, error) {
	defer t.machineExec.WsConnsLock.Unlock()
	t.machineExec.WsConnsLock.Lock()

	fmt.Println(" Output! " + string(p))
	for _, wsConn := range t.machineExec.WsConns {
		err := wsConn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			fmt.Println("Failed to write data " + err.Error())
		}
	}
	return len(p), nil
}

func (t PtyHandlerImpl) Next() *remotecommand.TerminalSize {
	log.Println("next size")
	select {
	case size := <-t.machineExec.SizeChan:
		log.Println("So next size will be ", size)
		return &size
	}
}
