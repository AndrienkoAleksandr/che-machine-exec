//
// Copyright (c) 2012-2018 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package model

import (
	"fmt"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"sync"
)

const (
	BufferSize = 8192
)

// todo get workspace id from env
type MachineIdentifier struct {
	MachineName string `json:"machineName"`
	WsId        string `json:"workspaceId"`
}

// todo machine exec can be the same for both: docker and kubernetes. Create separated
// session object to store sessionId and connection specific stuff
type MachineExec struct {
	Identifier MachineIdentifier `json:"identifier"`
	Cmd        []string          `json:"cmd"`
	Tty        bool              `json:"tty"`
	// Initial cols size
	Cols       int               `json:"cols"`
	// Initial rows size
	Rows       int               `json:"rows"`

	// unique client id, real execId should be hidden from client to prevent serialization
	ID int `json:"id"`
	// todo this is docker specific. Think where is it should be
	//ExecId string
	//Hjr    *types.HijackedResponse
	Buffer line_buffer.LineRingBuffer // move to InOutHandler

	//todo move it to the connectionHandler!!!
	WsConnsLock *sync.Mutex
	WsConns     []*websocket.Conn
	MsgChan     chan []byte

	//Started  bool
	Attached bool

	PtyHandler interface{} // here should be InOutHandler!!!
}

type InOutHandler interface {
	execIsAttached()
	Stream()
	// Restore
}

type InOutHandlerBase struct {
	InOutHandler

	MsgChan     chan []byte
	Buffer line_buffer.LineRingBuffer
}



func (machineExec *MachineExec) AddWebSocket(wsConn *websocket.Conn) {
	defer machineExec.WsConnsLock.Unlock()
	machineExec.WsConnsLock.Lock()

	machineExec.WsConns = append(machineExec.WsConns, wsConn)
}

func (machineExec *MachineExec) RemoveWebSocket(wsConn *websocket.Conn) {
	defer machineExec.WsConnsLock.Unlock()
	machineExec.WsConnsLock.Lock()

	for index, wsConnElem := range machineExec.WsConns {
		if wsConnElem == wsConn {
			machineExec.WsConns = append(machineExec.WsConns[:index], machineExec.WsConns[index+1:]...)
		}
	}
}

func (machineExec *MachineExec) getWSConns() []*websocket.Conn {
	defer machineExec.WsConnsLock.Unlock()
	machineExec.WsConnsLock.Lock()

	return machineExec.WsConns
}

//func (machineExec *MachineExec) Start() {
//	if machineExec.Hjr == nil {
//		return
//	}
//
//	go sendClientInputToExec(machineExec)
//	go sendExecOutputToWebsockets(machineExec)
//
//	machineExec.Started = true
//}
//
//func sendClientInputToExec(machineExec *MachineExec) {
//	for {
//		data := <-machineExec.MsgChan
//		if _, err := machineExec.Hjr.Conn.Write(data); err != nil {
//			fmt.Println("Failed to write data to exec with id ", machineExec.ID, " Cause: ", err.Error())
//			return
//		}
//	}
//}
//
//func sendExecOutputToWebsockets(machineExec *MachineExec) {
//	hjReader := machineExec.Hjr.Reader
//	buf := make([]byte, BufferSize)
//	var buffer bytes.Buffer
//
//	for {
//		rbSize, err := hjReader.Read(buf)
//		if err != nil {
//			//todo handle EOF error
//			fmt.Println("failed to read exec stdOut/stdError stream!!! " + err.Error())
//			return
//		}
//
//		i, err := normalizeBuffer(&buffer, buf, rbSize)
//		if err != nil {
//			log.Printf("Couldn't normalize byte buffer to UTF-8 sequence, due to an error: %s", err.Error())
//			return
//		}
//
//		if rbSize > 0 {
//			// save data to buffer to restore
//			//wsConns := machineExec.getWSConns()
//			//
//			//fmt.Println("Amount connections ", len(wsConns))
//
//			machineExec.WriteDataToWsConnections(buffer.Bytes())
//		}
//
//		buffer.Reset()
//		if i < rbSize {
//			buffer.Write(buf[i:rbSize])
//		}
//	}
//}

func (machineExec *MachineExec) WriteDataToWsConnections(data []byte) {
	defer machineExec.WsConnsLock.Unlock()
	machineExec.WsConnsLock.Lock()

	// save data to restore
	machineExec.Buffer.Write(data)
	// send data to the all connected clients
	for _, wsConn := range machineExec.WsConns {
		if err := wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("failed to write to ws-conn message!!!" + err.Error())
			machineExec.RemoveWebSocket(wsConn)
		}
	}
}
