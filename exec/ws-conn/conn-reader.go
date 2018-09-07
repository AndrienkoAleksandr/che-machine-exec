package ws_conn

import (
	"fmt"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"log"
)

func ReadWebSocketData(machineExec *model.MachineExec, wsConn *websocket.Conn) {
	defer machineExec.RemoveWebSocket(wsConn)

	for {
		msgType, wsBytes, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("failed to read ws-conn message") // todo better handle ws-conn error
			return
		}

		fmt.Println(" Message from client " + string(wsBytes))

		// todo check it. Seems unstable code here
		if msgType != websocket.TextMessage {
			continue
		}

		machineExec.MsgChan <- wsBytes
	}
}
