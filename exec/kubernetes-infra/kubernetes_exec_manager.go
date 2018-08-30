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

	"github.com/eclipse/che-machine-exec/api/model"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type MachineExecs struct {
	mutex   *sync.Mutex
	execMap map[int]*model.MachineExec
}

type KubernetesExecManager struct {
	client *kubernetes.Clientset
	// todo apply registry
}

// todo create exec registry to store list lanched execs.
// todo create client when we detected infra
var (
	config *rest.Config

	machineExecs = MachineExecs{
		mutex:   &sync.Mutex{},
		execMap: make(map[int]*model.MachineExec),
	}
	prevExecID uint64 = 0
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

func (KubernetesExecManager) Create(machineExec *model.MachineExec) (int, error) {
	return 0, nil
}

func (KubernetesExecManager) Check(id int) (int, error) {
	return 0, nil
}

func (KubernetesExecManager) Attach(id int) (*model.MachineExec, error) {
	return nil, nil
}

func (KubernetesExecManager) Resize(id int, cols uint, rows uint) error {
	return nil
}

//func getById(id int) *model.MachineExec {
//	defer machineExecs.mutex.Unlock()
//
//	machineExecs.mutex.Lock()
//	return machineExecs.execMap[id]
//}
