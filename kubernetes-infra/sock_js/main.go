package main

import (
	"log"
	"net/http"

	"fmt"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
)

func main() {
	handler := sockjs.NewHandler("/echo", sockjs.DefaultOptions, echoHandler)
	log.Fatal(http.ListenAndServe(":8081", handler))
}

func echoHandler(session sockjs.Session) {
	fmt.Println("id", session.ID())
	for {
		if msg, err := session.Recv(); err == nil {
			session.Send(msg)
			fmt.Println("id", session.ID())
			continue
		}
		break
	}
}
