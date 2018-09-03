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

package kubernetes_infra

import (
	"sync"
	"log"
	"errors"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"strconv"
	"sync/atomic"
	wsConnHandler "github.com/eclipse/che-machine-exec/exec/ws-conn"
)

//remove this after registry creation
type MachineExecs struct {
	mutex   *sync.Mutex
	execMap map[int]*model.MachineExec
}

type KubernetesExecManager struct {
	client *kubernetes.Clientset
	// todo apply registry
}

// todo create exec registry to store list launched execs.
var (
	config *rest.Config

	machineExecs = MachineExecs{
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*model.MachineExec),
	}
	prevExecID uint64 = 0
)

const (
	Pods = "pods"
	Exec = "exec"
	Post = "POST"
)

/**
 * Create new instance of the kubernetes exec manager
 */
func New() KubernetesExecManager {
	return KubernetesExecManager{client: createClient()}
}

func createClient() *kubernetes.Clientset {
	var err error

	//creates the in-cluster config
	//config, err = rest.InClusterConfig()
	//if err != nil {
	//	panic(err.Error())
	//}

	config, err = clientcmd.BuildConfigFromFlags("", "/home/user/.kube/config")
	if err != nil {
		glog.Fatal(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func (manager KubernetesExecManager) Create(machineExec *model.MachineExec) (int, error) {
	containerInfo, err := findMachineContainerInfo(manager, &machineExec.Identifier)

	req := manager.client.CoreV1().RESTClient().Post().
		Resource(Pods).
		Name(containerInfo.podName).
		Namespace(containerInfo.namespace).
		SubResource(Exec).
		// set up params
		VersionedParams(&v1.PodExecOptions{
			Container: containerInfo.name,
			Command:   machineExec.Cmd,
			Stdout:    true,
			Stderr:    true,
			Stdin:     true,
			TTY:       machineExec.Tty,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, Post, req.URL())
	if err != nil {
		return -1, err
	}

	defer machineExecs.mutex.Unlock()
	machineExecs.mutex.Lock()

	machineExec.Executor = executor
	machineExec.ID = int(atomic.AddUint64(&prevExecID, 1))
	machineExec.Buffer = line_buffer.CreateNewLineRingBuffer()
	machineExec.MsgChan = make(chan []byte)
	machineExec.WsConnsLock = &sync.Mutex{}
	machineExec.WsConns = make([]*websocket.Conn, 0)

	machineExecs.execMap[machineExec.ID] = machineExec

	log.Println("Create exec ", machineExec.ID, "execId", machineExec.ExecId)

	return machineExec.ID, nil
}

func (KubernetesExecManager) Check(id int) (int, error) {
	machineExec := getById(id)
	if machineExec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return machineExec.ID, nil
}

// todo made attach like void method...
func (KubernetesExecManager) Attach(id int, conn *websocket.Conn) (*model.MachineExec, error) {
	machineExec := getById(id)
	if machineExec == nil {
		return nil, errors.New("Exec '" + strconv.Itoa(id) + "' to attach was not found")
	}

	machineExec.AddWebSocket(conn)
	go wsConnHandler.ReadWebSocketData(machineExec, conn)
	go wsConnHandler.SendPingMessage(conn)

	// restore output...
	restoreContent := machineExec.Buffer.GetContent()
	conn.WriteMessage(websocket.TextMessage, []byte(restoreContent))

	ptyHandler := PtyHandlerImpl{machineExec: machineExec}

	err := machineExec.Executor.Stream(remotecommand.StreamOptions{
		Stdin:  ptyHandler,
		Stdout: ptyHandler,
		Stderr: ptyHandler,
		Tty:    machineExec.Tty,
	})
	if err != nil {
		return machineExec, err
	}

	return machineExec, nil
}

func (KubernetesExecManager) Resize(id int, cols uint, rows uint) error {
	return nil
}

func getById(id int) *model.MachineExec {
	defer machineExecs.mutex.Unlock()

	machineExecs.mutex.Lock()
	return machineExecs.execMap[id]
}
