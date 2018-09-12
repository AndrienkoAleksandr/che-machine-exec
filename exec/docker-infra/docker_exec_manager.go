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
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"golang.org/x/net/context"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
)

//remove this after registry creation
type DockerMachineExecManager struct {
	client *client.Client
	// todo apply registry
}

type MachineExecs struct {
	mutex   *sync.Mutex
	execMap map[int]*model.MachineExec
}

// todo create exec registry to store list launched execs.
var (
	execs = MachineExecs{
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*model.MachineExec),
	}
	prevExecID uint64 = 0
)

func NewManager() DockerMachineExecManager {
	return DockerMachineExecManager{client: createClient()}
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

func (manager DockerMachineExecManager) Create(exec *model.MachineExec) (int, error) {
	container, err := findMachineContainer(manager, &exec.Identifier)
	if err != nil {
		return -1, err
	}

	fmt.Println("found container for creation exec! id=", container.ID)

	resp, err := manager.client.ContainerExecCreate(context.Background(), container.ID, types.ExecConfig{
		Tty:          exec.Tty,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Detach:       false,
		Cmd:          exec.Cmd,
	})
	if err != nil {
		return -1, err
	}

	defer execs.mutex.Unlock()
	execs.mutex.Lock()

	exec.PtyHandler = NewPtyHandler(exec, resp.ID)

	exec.ID = int(atomic.AddUint64(&prevExecID, 1))
	exec.Buffer = line_buffer.CreateNewLineRingBuffer()


	execs.execMap[exec.ID] = exec

	fmt.Println("Create exec ", exec.ID, "execId", resp.ID)

	return exec.ID, nil
}

func (manager DockerMachineExecManager) Check(id int) (int, error) {
	exec := getById(id)
	if exec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return exec.ID, nil
}

func (manager DockerMachineExecManager) Attach(id int, conn *websocket.Conn) error {
	exec := getById(id)
	if exec == nil {
		return errors.New("Exec '" + strconv.Itoa(id) + "' to attach was not found")
	}

	// todo bad casting
	ptyHandler := exec.PtyHandler.(*DockerExecStreamHandler)

	ptyHandler.ConnsHandler.AddConnection(conn)
	go ptyHandler.ConnsHandler.ReadDataFromConnections(exec.PtyHandler, conn)
	go ptyHandler.ConnsHandler.SendPingMessage(conn)

	if exec.Attached {
		// restore previous output.
		log.Println("Restore content")
		restoreContent := exec.Buffer.GetContent()
		conn.WriteMessage(websocket.TextMessage, []byte(restoreContent))
		return nil
	}

	hjr, err := manager.client.ContainerExecAttach(context.Background(), ptyHandler.execId, types.ExecConfig{
		Detach: false,
		Tty:    exec.Tty,
	})
	if err != nil {
		return errors.New("Failed to attach to exec " + err.Error())
	}

	ptyHandler.hjr = &hjr
	exec.Attached = true

	ptyHandler.Stream()

	return nil
}

func (manager DockerMachineExecManager) Resize(id int, cols uint, rows uint) error {
	exec := getById(id)
	if exec == nil {
		return errors.New("Exec to resize '" + strconv.Itoa(id) + "' was not found")
	}

	ptyHandler := exec.PtyHandler.(*DockerExecStreamHandler)

	resizeParam := types.ResizeOptions{Height: rows, Width: cols}
	if err := manager.client.ContainerExecResize(context.Background(), ptyHandler.execId, resizeParam); err != nil {
		return err
	}

	return nil
}

func getById(id int) *model.MachineExec {
	defer execs.mutex.Unlock()

	execs.mutex.Lock()
	return execs.execMap[id]
}
