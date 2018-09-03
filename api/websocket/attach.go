package websocket

import (
	"net/http"
	"github.com/eclipse/che/agents/go-agents/core/rest"
	"strconv"
	"log"
	"github.com/eclipse/che-machine-exec/exec"
	"errors"
	"github.com/eclipse/che-lib/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func Attach(w http.ResponseWriter, r *http.Request, restParmas rest.Params) error {
	id, err := strconv.Atoi(restParmas.Get("id"))
	if err != nil {
		return errors.New("failed to parse id")
	}
	log.Println("Parsed id", id)

	wsConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Unable to upgrade connection to ws-conn " + err.Error())
		return err
	}

	if err = exec.GetExecManager().Attach(id, wsConn); err != nil {
		return err
	}

	return nil
}


