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
	"errors"
	"github.com/eclipse/che-lib/websocket"
	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/eclipse/che-machine-exec/line-buffer"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/golang/glog"
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

	execs = MachineExecs{
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
func NewManager() KubernetesExecManager {
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

func (manager KubernetesExecManager) Create(exec *model.MachineExec) (int, error) {
	log.Println("Begin creation")
	log.Println("Try to find container")
	containerInfo, err := findMachineContainerInfo(manager, &exec.Identifier)
	if err != nil {
		return -1, err
	}

	log.Println("Generate request")
	req := manager.client.CoreV1().RESTClient().Post().
		Resource(Pods).
		Name(containerInfo.podName).
		Namespace(containerInfo.namespace).
		SubResource(Exec).
		// set up params
		VersionedParams(&v1.PodExecOptions{
			Container: containerInfo.name,
			Command:   exec.Cmd,
			Stdout:    true,
			Stderr:    true,
			Stdin:     true,
			TTY:       exec.Tty,
		}, scheme.ParameterCodec)

	log.Println("Try to create new executor")
	executor, err := remotecommand.NewSPDYExecutor(config, Post, req.URL())
	if err != nil {
		return -1, err
	}

	log.Println("Created executor")
	defer execs.mutex.Unlock()
	execs.mutex.Lock()

	exec.ID = int(atomic.AddUint64(&prevExecID, 1))
	exec.Buffer = line_buffer.CreateNewLineRingBuffer()

	exec.PtyHandler = NewPtyHandler(exec, executor)

	execs.execMap[exec.ID] = exec

	log.Println("Create exec ", exec.ID, "execId", exec.ID)

	return exec.ID, nil
}

func (KubernetesExecManager) Check(id int) (int, error) {
	exec := getById(id)
	if exec == nil {
		return -1, errors.New("Exec '" + strconv.Itoa(id) + "' was not found")
	}
	return exec.ID, nil
}

func (KubernetesExecManager) Attach(id int, conn *websocket.Conn) error {
	exec := getById(id)
	if exec == nil {
		return errors.New("Exec '" + strconv.Itoa(id) + "' to attach was not found")
	}

	// todo bad casting
	ptyHandler := exec.PtyHandler.(*KubernetesExecStreamHandler)

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


	exec.Attached = true

	return ptyHandler.Stream(exec.Tty)

	//err := ptyHandler.executor.Stream(remotecommand.StreamOptions{
	//	Stdin:             ptyHandler,
	//	Stdout:            ptyHandler,
	//	Stderr:            ptyHandler,
	//	TerminalSizeQueue: ptyHandler,
	//	Tty:               exec.Tty,
	//})
	//if err != nil {
	//	return err
	//}
}

func (KubernetesExecManager) Resize(id int, cols uint, rows uint) error {
	log.Println("Resize!!!")
	exec := getById(id)
	if exec == nil {
		return errors.New("Exec to resize '" + strconv.Itoa(id) + "' was not found")
	}

	ptyHandler := exec.PtyHandler.(*KubernetesExecStreamHandler)

	log.Println("take a look on the chan", ptyHandler.sizeChan)

	ptyHandler.sizeChan <- remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	return nil
}

func getById(id int) *model.MachineExec {
	defer execs.mutex.Unlock()

	execs.mutex.Lock()
	return execs.execMap[id]
}
