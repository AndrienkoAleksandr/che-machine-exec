package docker_infra

import (
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"fmt"
	"bytes"
	"log"
)

type DockerPtyHandler struct {
	exec *model.MachineExec

	execId string
	hjr    *types.HijackedResponse
}

func NewPtyHandler(exec *model.MachineExec, execId string) *DockerPtyHandler {
	return &DockerPtyHandler{ exec: exec, execId:execId}
}

func (ptyH DockerPtyHandler) Stream() {
	if ptyH.hjr == nil {
		return
	}

	go ptyH.sendClientInputToExec()
	go ptyH.sendExecOutputToWebsockets()
}

func (ptyH DockerPtyHandler) sendClientInputToExec() {
	machineExec := ptyH.exec
	for {
		data := <-machineExec.MsgChan // todo move MsgChan to pty!!!!
		if _, err := ptyH.hjr.Conn.Write(data); err != nil {
			fmt.Println("Failed to write data to exec with id ", machineExec.ID, " Cause: ", err.Error())
			return
		}
	}
}

func (ptyH DockerPtyHandler) sendExecOutputToWebsockets() {
	hjReader := ptyH.hjr.Reader
	buf := make([]byte, model.BufferSize)
	var buffer bytes.Buffer

	for {
		rbSize, err := hjReader.Read(buf)
		if err != nil {
			//todo handle EOF error
			fmt.Println("failed to read exec stdOut/stdError stream!!! " + err.Error())
			return
		}

		i, err := model.NormalizeBuffer(&buffer, buf, rbSize)
		if err != nil {
			log.Printf("Couldn't normalize byte buffer to UTF-8 sequence, due to an error: %s", err.Error())
			return
		}

		if rbSize > 0 {
			ptyH.exec.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}

