package server

import (
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"k8s.io/client-go/tools/remotecommand"
	"bytes"
	"fmt"

	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/gorilla/websocket"
	"log"
	"sync"
)

const BufferSize = 8192

type ServerExec struct { // todo rename... ExecSession?
	*model.MachineExec

	// Todo Refactoring this code is docker specific. Create separated code layer and move it.
	ExecId string
	Hjr    *types.HijackedResponse

	// Todo Refactoring: this code is websocket connection specific. Move this code.
	WsConnsLock *sync.Mutex
	WsConns     []*websocket.Conn
	MsgChan     chan []byte

	// Todo Refactoring: this code is kubernetes specific. Create separated code layer and move it.
	Executor remotecommand.Executor
	SizeChan chan remotecommand.TerminalSize

	// Todo Refactoring: Create separated code layer and move it.
	Buffer *line_buffer.LineRingBuffer
}

// split kubernetes and docker specific logic to the ServerKubernetesExec and DockerKubernetesExec based on ServerExec
func NewServerExec(machineExec *model.MachineExec, execId string, executor remotecommand.Executor) *ServerExec  {
	return &ServerExec{
		MachineExec: machineExec,
		ExecId: execId,
		MsgChan: make(chan []byte),
		WsConnsLock: &sync.Mutex{},
		WsConns: make([]*websocket.Conn, 0),
		Executor: executor,
		SizeChan : make(chan remotecommand.TerminalSize),
	}
}

func (machineExec *ServerExec) AddWebSocket(wsConn *websocket.Conn) {
	defer machineExec.WsConnsLock.Unlock()
	machineExec.WsConnsLock.Lock()

	machineExec.WsConns = append(machineExec.WsConns, wsConn)
}

func (exec *ServerExec) RemoveWebSocket(wsConn *websocket.Conn) {
	defer exec.WsConnsLock.Unlock()
	exec.WsConnsLock.Lock()

	for index, wsConnElem := range exec.WsConns {
		if wsConnElem == wsConn {
			exec.WsConns = append(exec.WsConns[:index], exec.WsConns[index+1:]...)
		}
	}
}

func (exec *ServerExec) getWSConns() []*websocket.Conn {
	defer exec.WsConnsLock.Unlock()
	exec.WsConnsLock.Lock()

	return exec.WsConns
}

func (machineExec *ServerExec) Start() {
	if machineExec.Hjr == nil {
		return
	}

	go sendClientInputToExec(machineExec)
	go sendExecOutputToWebsockets(machineExec)
}

func sendClientInputToExec(machineExec *ServerExec) {
	for {
		data := <-machineExec.MsgChan
		if _, err := machineExec.Hjr.Conn.Write(data); err != nil {
			fmt.Println("Failed to write data to exec with id ", machineExec.ID, " Cause: ", err.Error())
			return
		}
	}
}

func sendExecOutputToWebsockets(machineExec *ServerExec) {
	hjReader := machineExec.Hjr.Reader
	buf := make([]byte, BufferSize)
	var buffer bytes.Buffer

	for {
		rbSize, err := hjReader.Read(buf)
		if err != nil {
			//todo handle EOF error
			fmt.Println("failed to read exec stdOut/stdError stream!!! " + err.Error())
			return
		}

		i, err := normalizeBuffer(&buffer, buf, rbSize)
		if err != nil {
			log.Printf("Couldn't normalize byte buffer to UTF-8 sequence, due to an error: %s", err.Error())
			return
		}

		if rbSize > 0 {
			machineExec.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}

func (exec *ServerExec) WriteDataToWsConnections(data []byte) {
	defer exec.WsConnsLock.Unlock()
	exec.WsConnsLock.Lock()

	// save data to restore
	exec.Buffer.Write(data)
	// send data to the all connected clients
	for _, wsConn := range exec.WsConns {
		if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("failed to write to ws-conn message!!!" + err.Error())
			exec.RemoveWebSocket(wsConn)
		}
	}
}