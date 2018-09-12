package docker_infra

import (
	"github.com/docker/docker/api/types"
	"github.com/eclipse/che-machine-exec/api/model"
	"fmt"
	"bytes"
	"log"
	"github.com/eclipse/che-machine-exec/api/websocket/ws-conn"
)

type DockerExecStreamHandler struct {
	*model.InOutHandlerBase

	execId string
	hjr    *types.HijackedResponse
}

func NewPtyHandler(execId string) *DockerExecStreamHandler {
	msgChan := make(chan []byte)
	connsHandler := ws_conn.New()
	inOutHandler := &model.InOutHandlerBase{MsgChan:msgChan, ConnsHandler: connsHandler}

	return &DockerExecStreamHandler{
		execId:execId,
		InOutHandlerBase:inOutHandler,
	}
}

func (strH DockerExecStreamHandler) Stream(tty bool) error {
	if strH.hjr == nil {
		return nil // todo create and return err!!!
	}

	go strH.sendClientInputToExec()
	go strH.sendExecOutputToWebSockets()

	return nil
}

func (strH DockerExecStreamHandler) sendClientInputToExec() {
	for {
		data := <-strH.MsgChan
		if _, err := strH.hjr.Conn.Write(data); err != nil {
			//log error!!! with machine id someHow...
			return
		}
	}
}

func (strH DockerExecStreamHandler) sendExecOutputToWebSockets() {
	hjReader := strH.hjr.Reader
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
			strH.Buffer.Write(buffer.Bytes())
			strH.ConnsHandler.WriteDataToWsConnections(buffer.Bytes())
		}

		buffer.Reset()
		if i < rbSize {
			buffer.Write(buf[i:rbSize])
		}
	}
}

