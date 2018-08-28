/*
Copyright 2016 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Note: the example only works with the code within the same release/branch.
package main

import (
	"fmt"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/remotecommand"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"encoding/json"
	"net/http"
	"github.com/eclipse/che-machine-exec/kubernetes-infra/pty"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/kubernetes/scheme"
	"encoding/hex"
	"math/rand"
)

var (
	namespace = "default"
	podName = "test-exec"
	containerName = "exec"
	cmd = []string{"/bin/sh"}
	client *kubernetes.Clientset
	config *rest.Config
)


func CreateAttachHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, handleTerminalSession)
}

// terminalSessions stores a map of all TerminalSession objects
// FIXME: this structure needs locking
var terminalSessions = make(map[string]pty.TerminalSession)

// handleTerminalSession is Called by net/http for any new /api/sockjs connections
func handleTerminalSession(session sockjs.Session) {
	fmt.Println("What")
	var (
		buf             string
		err             error
		msg             pty.TerminalMessage
		terminalSession pty.TerminalSession
		ok              bool
	)

	if buf, err = session.Recv(); err != nil {
		fmt.Printf("handleTerminalSession: can't Recv: %v", err)
		return
	}

	if err = json.Unmarshal([]byte(buf), &msg); err != nil {
		fmt.Printf("handleTerminalSession: can't UnMarshal (%v): %s", err, buf)
		return
	}

	if msg.Op != "bind" {
		fmt.Printf("handleTerminalSession: expected 'bind' message, got: %s", buf)
		return
	}

	if terminalSession, ok = terminalSessions[msg.SessionID]; !ok {
		fmt.Printf("handleTerminalSession: can't find session '%s'", msg.SessionID)
		return
	}

	terminalSession.SockJSSession = session
	terminalSessions[msg.SessionID] = terminalSession
	terminalSession.Bound <- nil
}

func createClient() *kubernetes.Clientset {
	var err error

	//creates the in-cluster config
	config, err = rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	//config, err = clientcmd.BuildConfigFromFlags("", "/home/user/.kube/config")
	//if err != nil {
	//	glog.Fatal(err)
	//}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

// genTerminalSessionId generates a random session ID string. The format is not really interesting.
// This ID is used to identify the session when the client opens the SockJS connection.
// Not the same as the SockJS session id! We can't use that as that is generated
// on the client side and we don't have it yet at this point.
func genTerminalSessionId() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := make([]byte, hex.EncodedLen(len(bytes)))
	hex.Encode(id, bytes)
	fmt.Println("Generated id" + string(id))
	return string(id), nil
}

// WaitForTerminal is called from apihandler.handleAttach as a goroutine
// Waits for the SockJS connection to be opened by the client the session to be bound in handleTerminalSession
func WaitForTerminal(k8sClient kubernetes.Interface, cfg *rest.Config, sessionId string) {
	shell := "/bin/bash"// request.QueryParameter("shell")

	select {
	case <-terminalSessions[sessionId].Bound:
		close(terminalSessions[sessionId].Bound)

		var err error

		cmd := []string{shell}
		err = startProcess(k8sClient, cfg, cmd, terminalSessions[sessionId])

		if err != nil {
			terminalSessions[sessionId].Close(2, err.Error())
			return
		}

		terminalSessions[sessionId].Close(1, "Process exited")
	}
}

func startProcess(k8sClient kubernetes.Interface, cfg *rest.Config, cmd []string, ptyHandler pty.PtyHandler) error {
	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Container: containerName,
		Command:   cmd,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(cfg, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:             ptyHandler,
		Stdout:            ptyHandler,
		Stderr:            ptyHandler,
		TerminalSizeQueue: ptyHandler,
		Tty:               true,
	})
	if err != nil {
		return err
	}

	return nil
}

// Handles execute shell API call
func handleExec(response http.ResponseWriter, request *http.Request) {
	sessionId, err := genTerminalSessionId()
	if err != nil {
		response.WriteHeader(500)
		return
	}

	terminalSessions[sessionId] = pty.TerminalSession{
		Id:       sessionId,
		Bound:    make(chan error),
		SizeChan: make(chan remotecommand.TerminalSize),
	}
	go WaitForTerminal(client, config, sessionId)
	if origin := request.Header.Get("Origin"); origin != "" {
		response.Header().Set("Access-Control-Allow-Origin", origin)
		response.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		response.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	}
	response.Header().Set("Session-Id", sessionId)
	response.Write([]byte(sessionId))
	//response.WriteHeader(http.StatusOK)
}

func main() {
	client = createClient()

	http.Handle("/api/sockjs/", CreateAttachHandler("/api/sockjs"))

	http.HandleFunc("/api/create", handleExec)
	//http.Handle("/api/sockjs/", ))

	url := ":4000"

	server := &http.Server{
		Addr:      url,
		Handler:   http.DefaultServeMux,
		//TLSConfig: &tls.Config{Certificates: servingCerts},
	}

	//if err := http.ListenAndServe(url, nil); err != nil {
	//	panic("Unable to connect!" + err.Error())
	//}
	server.ListenAndServe()
}
