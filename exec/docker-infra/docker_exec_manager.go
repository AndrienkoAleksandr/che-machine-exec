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
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"log"
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

	exec := NewDockerExecSession(machineExec, resp.ID)

	return manager.registry.Add(exec), nil
}

func (manager DockerMachineExecManager) Check(id int) (int, error) {
	exec, err := manager.getSessionById(id)
	if err != nil {
		return -1, err
	}
	return exec.Id(), nil
}

func (manager DockerMachineExecManager) Attach(id int, conn *websocket.Conn) error {
	exec, err := manager.getSessionById(id)
	if err != nil {
		return err
	}

	exec.ConnHandler.AddConnection(conn)
	go exec.ConnHandler.ReadDataFromConnections(exec.MsgChan, conn)
	go exec.ConnHandler.SendPingMessage(conn)

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

	exec.Stream()

	return nil
}

func (manager DockerMachineExecManager) Resize(id int, cols uint, rows uint) error {
	exec, err := manager.getSessionById(id)
	if err != nil {
		return err
	}

	resizeParam := types.ResizeOptions{Height: rows, Width: cols}
	if err := manager.client.ContainerExecResize(context.Background(), exec.ExecId, resizeParam); err != nil {
		return err
	}

	return nil
}

func (manager DockerMachineExecManager) getSessionById(id int) (*DockerExecSession, error) {
	execSession := manager.registry.GetById(id)

	if execSession == nil {
		log.Println("Exec session was not found for exec with id: " + strconv.Itoa(id))
		return nil, errors.New("Exec session was not found for exec with id: " + strconv.Itoa(id))
	}

	dSession, ok := execSession.(*DockerExecSession)
	if !ok {
		log.Println("Stored incompatible exec session type for docker infrastructure. Exec id: " + strconv.Itoa(id))
		return nil, errors.New("stored incompatible exec session type for docker infrastructure. Exec id: " + strconv.Itoa(id))
	}

	return dSession, nil
}
