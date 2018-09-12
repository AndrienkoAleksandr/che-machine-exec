package ws_conn

import (
	"fmt"
	"github.com/eclipse/che-lib/websocket"
	"log"
	"time"
	"sync"
	"github.com/eclipse/che-machine-exec/api/temp"
)

const PingPeriod = 30 * time.Second

type ConnectionHandler struct {
	// todo check incupsulation and package scope
	WsConnsLock *sync.Mutex
	WsConns     []*websocket.Conn
}

func New() *ConnectionHandler {
	return &ConnectionHandler{
		WsConnsLock: &sync.Mutex{},
		WsConns: make([]*websocket.Conn, 0),
	}
}

func (connHandler *ConnectionHandler) AddConnection(wsConn *websocket.Conn) {
	fmt.Println("Add websocket")
	defer connHandler.WsConnsLock.Unlock()
	connHandler.WsConnsLock.Lock()

	connHandler.WsConns = append(connHandler.WsConns, wsConn)
}

func (connHandler *ConnectionHandler) RemoveWebSocket(wsConn *websocket.Conn) {
	fmt.Println("Remove websocket")
	defer connHandler.WsConnsLock.Unlock()
	connHandler.WsConnsLock.Lock()

	for index, wsConnElem := range connHandler.WsConns {
		if wsConnElem == wsConn {
			connHandler.WsConns = append(connHandler.WsConns[:index], connHandler.WsConns[index+1:]...)
		}
	}
}

// can be usefull for testing purpose...
//func (connHandler *ConnectionHandler) getWSConns() []*websocket.Conn {
//	defer connHandler.WsConnsLock.Unlock()
//	connHandler.WsConnsLock.Lock()
//
//	return connHandler.WsConns
//}

func (connHandler *ConnectionHandler) WriteDataToWsConnections(data []byte) {
	defer connHandler.WsConnsLock.Unlock()
	connHandler.WsConnsLock.Lock()

	// save data to restore
	// todo restore is broken!!!
	//connHandler.Buffer.Write(data)
	// send data to the all connected clients
	fmt.Println("ws connection size", len(connHandler.WsConns))
	for _, wsConn := range connHandler.WsConns {
		if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("failed to write to ws-conn message!!!" + err.Error())
			connHandler.RemoveWebSocket(wsConn)
		}
	}
}

//todo use here pty handler instead of io.writer?
func (connHandler *ConnectionHandler) ReadDataFromConnections(machineExecWriter temp.StreamWriter, wsConn *websocket.Conn) {
	defer connHandler.RemoveWebSocket(wsConn)

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

		machineExecWriter.WriteInput(wsBytes)
	}
}

// can be idependent method
func (connHandler *ConnectionHandler) SendPingMessage(wsConn *websocket.Conn) {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			log.Printf("Error occurs on sending ping message to ws-conn. %v", err)
			return
		}
	}
}