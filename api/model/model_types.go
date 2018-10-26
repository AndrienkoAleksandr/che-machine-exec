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

package model

// todo inside workspace we can get workspace id from env variables.
type MachineIdentifier struct {
	MachineName string `json:"machineName"`
	WsId        string `json:"workspaceId"`
}

type MachineExec struct {
	Identifier MachineIdentifier `json:"identifier"`
	Cmd        []string          `json:"cmd"`
	Tty        bool              `json:"tty"`
	Cols       int               `json:"cols"`
	Rows       int               `json:"rows"`

	// unique client id, real execId should be hidden from client to prevent serialization
	ID int `json:"id"`
}
