package transport

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

const PingPeriod = 30 * time.Second

type ConnectionHandler struct {
	wsConnsLock *sync.Mutex
	wsConns     []*websocket.Conn
}

func NewConnHandler() *ConnectionHandler {
	return &ConnectionHandler{
		wsConnsLock: &sync.Mutex{},
		wsConns:     make([]*websocket.Conn, 0),
	}
}

func (handler *ConnectionHandler) AddConnection(wsConn *websocket.Conn) {
	fmt.Println("Add websocket")
	defer handler.wsConnsLock.Unlock()
	handler.wsConnsLock.Lock()

	handler.wsConns = append(handler.wsConns, wsConn)
}

func (handler *ConnectionHandler) RemoveConnection(wsConn *websocket.Conn) {
	fmt.Println("Remove websocket")
	defer handler.wsConnsLock.Unlock()
	handler.wsConnsLock.Lock()

	for index, wsConnElem := range handler.wsConns {
		if wsConnElem == wsConn {
			handler.wsConns = append(handler.wsConns[:index], handler.wsConns[index+1:]...)
		}
	}
}

func (handler *ConnectionHandler) WriteDataToWsConnections(data []byte) {
	defer handler.wsConnsLock.Unlock()
	handler.wsConnsLock.Lock()

	fmt.Println("ws connection size", len(handler.wsConns))
	for _, wsConn := range handler.wsConns {
		if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("failed to write to ws-conn message!!!" + err.Error())
			handler.RemoveConnection(wsConn)
		}
	}
}

func (handler *ConnectionHandler) ReadDataFromConnections(inputChan chan []byte, wsConn *websocket.Conn) {
	defer handler.RemoveConnection(wsConn)

	for {
		msgType, wsBytes, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("failed to read ws-conn message")
			return
		}

		fmt.Println(" Message from client " + string(wsBytes))

		if msgType != websocket.TextMessage {
			continue
		}

		inputChan <- wsBytes
	}
}

func (*ConnectionHandler) SendPingMessage(wsConn *websocket.Conn) {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			log.Printf("Error occurs on sending ping message to ws-conn. %v", err)
			return
		}
	}
}