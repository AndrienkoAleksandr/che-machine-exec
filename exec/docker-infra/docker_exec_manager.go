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

package docker_infra

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/exec/registry"
	"github.com/eclipse/che-machine-exec/exec/server"
	wsConnHandler "github.com/eclipse/che-machine-exec/exec/ws-conn"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"strconv"
)

type DockerMachineExecManager struct {
	client *client.Client
	registry *registry.ExecRegistry
}

func New() DockerMachineExecManager {
	return DockerMachineExecManager{
		client: createClient(),
		registry: registry.NewExecRegistry(),
	}
}

func createClient() *client.Client {
	dockerClient, err := client.NewEnvClient()
	// set up minimal docker version 1.13.0(api version 1.25).
	dockerClient.UpdateClientVersion("1.25")
	if err != nil {
		panic(err)
	}
	return dockerClient
}

func (manager DockerMachineExecManager) Create(machineExec *model.MachineExec) (int, error) {
	container, err := findMachineContainer(manager, &machineExec.Identifier)
	if err != nil {
		return -1, err
	}

	fmt.Println("found container for creation exec! id=", container.ID)

	resp, err := manager.client.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
		Tty:          machineExec.Tty,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		Cmd:          machineExec.Cmd,
	})
	if err != nil {
		return -1, err
	}

	exec := server.NewServerExec(machineExec, resp.ID, nil)

	return manager.registry.Add(exec), nil
}

func (manager DockerMachineExecManager) Check(id int) (int, error) {
	exec := manager.registry.GetById(id)
	if exec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return exec.ID, nil
}

func (manager DockerMachineExecManager) Attach(id int, conn *websocket.Conn) error {
	exec := manager.registry.GetById(id)
	if exec == nil {
		return errors.New("Exec '" + strconv.Itoa(id) + "' to attach was not found")
	}

	exec.AddWebSocket(conn)
	go wsConnHandler.ReadWebSocketData(exec, conn)
	go wsConnHandler.SendPingMessage(conn)

	if exec.Buffer != nil {
		// restore previous output.
		restoreContent := exec.Buffer.GetContent()
		return conn.WriteMessage(websocket.TextMessage, []byte(restoreContent))
	}

	hjr, err := manager.client.ContainerExecAttach(context.Background(), exec.ExecId, types.ExecConfig{
		Detach: false,
		Tty:    exec.Tty,
	})
	if err != nil {
		return errors.New("Failed to attach to exec " + err.Error())
	}

	exec.Hjr = &hjr
	exec.Buffer = line_buffer.New()

	exec.Start()

	return nil
}

func (manager DockerMachineExecManager) Resize(id int, cols uint, rows uint) error {
	exec := manager.registry.GetById(id)
	if exec == nil {
		return errors.New("Exec to resize '" + strconv.Itoa(id) + "' was not found")
	}

	resizeParam := types.ResizeOptions{Height: rows, Width: cols}
	if err := manager.client.ContainerExecResize(context.Background(), exec.ExecId, resizeParam); err != nil {
		return err
	}

	return nil
}
