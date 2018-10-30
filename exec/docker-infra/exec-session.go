package docker_infra

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/exec/server"
	"github.com/pkg/errors"
	"log"
)

const BufferSize = 8192

type DockerExecSession struct {
	*server.ExecSessionBase

	ExecId string
	Hjr    *types.HijackedResponse
}

func NewDockerExecSession(machineExec *model.MachineExec, ExecId string) *DockerExecSession {
	return &DockerExecSession{
		ExecSessionBase: server.NewExecSessionBase(machineExec),
		ExecId: ExecId,
	}
}

func (execSession *DockerExecSession) Stream() error {
	if execSession.Hjr == nil {
		return errors.New("Exec is not attached yet.")
	}

	go sendClientInputToExec(execSession)
	go sendExecOutputToWebSockets(execSession)

	return nil
}

func sendClientInputToExec(exec *DockerExecSession) {
	for {
		data := <-exec.MsgChan
		if _, err := exec.Hjr.Conn.Write(data); err != nil {
			fmt.Println("Failed to write data to exec with id ", exec.ID, " Cause: ", err.Error())
			return
		}
	}
}

func sendExecOutputToWebSockets(exec *DockerExecSession) {
	hjReader := exec.Hjr.Reader
	buf := make([]byte, BufferSize)
	var buffer bytes.Buffer

	for {
		rbSize, err := hjReader.Read(buf)
		if err != nil {
			fmt.Println("failed to read exec stdOut/stdError stream!!! " + err.Error())
			return
		}

		i, err := server.NormalizeBuffer(&buffer, buf, rbSize)
		if err != nil {
			log.Printf("Couldn't normalize byte buffer to UTF-8 sequence, due to an error: %s", err.Error())
			return
		}

		if rbSize > 0 {
			// save data to restore
			exec.Buffer.Write(buffer.Bytes())
			exec.ConnHandler.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}