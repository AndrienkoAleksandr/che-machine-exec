package server

import (
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/transport"
	"k8s.io/client-go/tools/remotecommand"
	"bytes"
	"fmt"

	"github.com/eclipse/che-machine-exec/line-buffer"
	"log"
)

const BufferSize = 8192

type ExecSession struct {
	*model.MachineExec

	// Todo Refactoring this code is docker specific. Create separated code layer and move it.
	ExecId string
	Hjr    *types.HijackedResponse

	ConnHandler *transport.ConnectionHandler

	MsgChan     chan []byte

	// Todo Refactoring: this code is kubernetes specific. Create separated code layer and move it.
	Executor remotecommand.Executor
	SizeChan chan remotecommand.TerminalSize

	// Todo Refactoring: Create separated code layer and move it.
	Buffer *line_buffer.LineRingBuffer
}

// split kubernetes and docker specific logic to the ServerKubernetesExec and DockerKubernetesExec based on ExecSession
func NewServerExec(machineExec *model.MachineExec, execId string, executor remotecommand.Executor) *ExecSession {
	return &ExecSession{
		MachineExec: machineExec,
		ExecId: execId,
		MsgChan: make(chan []byte),
		Executor: executor,
		ConnHandler: transport.NewConnHandler(),
		SizeChan : make(chan remotecommand.TerminalSize),
	}
}

func (exec *ExecSession) Start() {
	if exec.Hjr == nil {
		return
	}

	go sendClientInputToExec(exec)
	go sendExecOutputToWebsockets(exec)
}

func sendClientInputToExec(exec *ExecSession) {
	for {
		data := <-exec.MsgChan
		if _, err := exec.Hjr.Conn.Write(data); err != nil {
			fmt.Println("Failed to write data to exec with id ", exec.ID, " Cause: ", err.Error())
			return
		}
	}
}

func sendExecOutputToWebsockets(exec *ExecSession) {
	hjReader := exec.Hjr.Reader
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
			exec.ConnHandler.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}
