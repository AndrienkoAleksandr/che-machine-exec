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
	"fmt"
	"github.com/eclipse/che-machine-exec/api/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	WsId        = "che.workspace_id"
	MachineName = "CHE_MACHINE_NAME"
)

// Find container name by pod label: "wsId" and container environment variables "machineName".
func findMachineContainer(execManager KubernetesExecManager, identifier *model.MachineIdentifier) (string, error) {
	pods, err := execManager.client.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: WsId + "=" + identifier.WsId})
	if err != nil {
		return "", err
	}
	containers := pods.Items[0].Spec.Containers

	var containerName string
	for _, container := range containers {
		for _, env := range container.Env {
			if env.Name == MachineName && env.Value == identifier.MachineName {
				containerName = container.Name
			}
		}
	}

	fmt.Println("Found container with name " + containerName)

	return containerName, nil
}
