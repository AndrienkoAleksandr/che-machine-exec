package server

import (
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/eclipse/che-machine-exec/transport"
)

type ExecSession interface {
	Id() int
	SetId(id int)

	Stream() error
}

type ExecSessionBase struct {
	ExecSession
	*model.MachineExec

	ConnHandler *transport.ConnectionHandler
	MsgChan     chan []byte
	Buffer *line_buffer.LineRingBuffer
}

func (session ExecSessionBase) Id() int {
	return session.ID
}

func (session ExecSessionBase) SetId(id int) {
	session.ID = id
}

func NewExecSessionBase(machineExec *model.MachineExec) *ExecSessionBase {
	return &ExecSessionBase{
		MachineExec: machineExec,
		MsgChan: make(chan []byte),
		ConnHandler: transport.NewConnHandler(),
	}
}
