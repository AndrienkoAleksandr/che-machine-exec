package ws_conn

import (
	"time"
	"github.com/eclipse/che-lib/websocket"
	"log"
)

const PingPeriod = 30 * time.Second

func SendPingMessage(wsConn *websocket.Conn) {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if err := wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
			log.Printf("Error occurs on sending ping message to ws-conn. %v", err)
			return
		}
	}
}


