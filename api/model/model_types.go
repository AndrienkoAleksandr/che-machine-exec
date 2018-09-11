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
	// unique machine exec id
	ID int `json:"id"`
	Identifier MachineIdentifier `json:"identifier"`
	Cmd        []string          `json:"cmd"`
	Tty        bool              `json:"tty"`
	// Initial cols size
	Cols       int               `json:"cols"`
	// Initial rows size
	Rows       int               `json:"rows"`


	Buffer line_buffer.LineRingBuffer // move to InOutHandler

	//todo move it to the connectionHandler!!!
	WsConnsLock *sync.Mutex
	WsConns     []*websocket.Conn
	// MsgChan     chan []byte

	//Started  bool
	Attached bool

	PtyHandler InOutHandler // here should be InOutHandler!!!
}

type InOutHandler interface {
	execIsAttached() bool
	//StreamOutput()
	WriteInput([]byte)
	//Resize(int cols, int rows)

	// Restore
}

type InOutHandlerBase struct {
	InOutHandler

	MsgChan     chan []byte
	//Buffer line_buffer.LineRingBuffer
}

// todo change to write and 'io' interface!!!
func (baseHandler InOutHandlerBase) WriteInput(bts []byte)  {
	baseHandler.MsgChan <- bts
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
