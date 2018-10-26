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

package ws_conn

import (
	"github.com/eclipse/che-machine-exec/exec/server"
	"github.com/gorilla/websocket"
	"log"
)

func ReadWebSocketData(exec *server.ServerExec, wsConn *websocket.Conn) {
	defer exec.RemoveWebSocket(wsConn)

	for {
		msgType, wsBytes, err := wsConn.ReadMessage()
		if err != nil {
			log.Printf("failed to read ws-conn message") // todo better handle ws-conn error
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		exec.MsgChan <- wsBytes
	}
}
