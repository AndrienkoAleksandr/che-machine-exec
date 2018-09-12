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
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/eclipse/che-machine-exec/api/websocket/ws-conn"
	"fmt"
	"github.com/eclipse/che-machine-exec/api/temp"
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

	// remove it!!! Implement and use isAttached
	Attached bool
	PtyHandler        InOutHandler // here should be InOutHandler!!!
}

//type ExecSessionStorage struct {
//	MachineExec *MachineExec
//	PtyHandler        InOutHandler
//}

// rename to StdInOutHandler
type InOutHandler interface {
	execIsAttached() bool

	temp.StreamWriter

	//StreamExecStdInOut()
	//io.Writer
	//Write([]byte) (int, error) // io.Writer....
	//Resize(int cols, int rows)
	// Restore
}

type InOutHandlerBase struct {
	InOutHandler



	ConnsHandler *ws_conn.ConnectionHandler

	MsgChan     chan []byte // rename outputChan
	// InputChan create it and use for ws connnections!!!!

	//Buffer line_buffer.LineRingBuffer
}

// todo change to write and 'io' interface!!!
func (baseHandler InOutHandlerBase) WriteInput(bts []byte)   {
	fmt.Println("Write here....")
	baseHandler.MsgChan <- bts
}




