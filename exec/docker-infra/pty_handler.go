package docker_infra

import (
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"fmt"
	"bytes"
	"log"
	"github.com/eclipse/che-machine-exec/api/websocket/ws-conn"
)

type DockerPtyHandler struct {
	*model.InOutHandlerBase

	// todo remove exec
	exec *model.MachineExec

	execId string
	hjr    *types.HijackedResponse
}

func NewPtyHandler(exec *model.MachineExec, execId string) *DockerPtyHandler {
	msgChan := make(chan []byte)
	connsHandler := ws_conn.New()
	inOutHandler := &model.InOutHandlerBase{MsgChan:msgChan, ConnsHandler: connsHandler}

	return &DockerPtyHandler{
		exec: exec,
		execId:execId,
		InOutHandlerBase:inOutHandler,
	}
}

func (ptyH DockerPtyHandler) Stream() {
	if ptyH.hjr == nil {
		return
	}

	go ptyH.sendClientInputToExec()
	go ptyH.sendExecOutputToWebSockets()
}

func (ptyH DockerPtyHandler) execIsAttached() bool {
	return false;
}

func (ptyH DockerPtyHandler) sendClientInputToExec() {
	for {
		data := <-ptyH.MsgChan
		if _, err := ptyH.hjr.Conn.Write(data); err != nil {
			//log error!!! with machine id someHow...
			return
		}
	}
}

func (ptyH DockerPtyHandler) sendExecOutputToWebSockets() {
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
			fmt.Println("send Data to the all connections!!!! " + string(buffer.Bytes()))
			ptyH.ConnsHandler.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}

