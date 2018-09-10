# CHE macine exec

Go-lang server side to creation machine-execs for Eclipse CHE workspaces.
Uses to spawn terminal or command processes.

CHE machine exec uses json-rpc protocol to communication with client.

# How to use machine-exec image with Eclipse CHE workspace on the docker infrastructure:
Apply docker.sock path (by default it's `/var/run/docker.sock`) to the workspace volume property `CHE_WORKSPACE_VOLUME` in the che.env file:
Example:
 ```
CHE_WORKSPACE_VOLUME=/var/run/docker.sock
```
che.env file located in the CHE `data` folder. che.env file contains configuration properties for Eclipse CHE. All changes of the file become avaliable after restart Eclipse CHE.
 
# Build docker image

Build docker image with che-machine-exec manually:

```
docker build -t eclipse/che-machine-exec .
```

# Run docker container

Run docker container with che-machine-exec manually:

```
docker run --rm -p 4444:4444 -v /var/run/docker.sock:/var/run/docker.sock eclipse/che-machine-exec
````

Create websocket connection to the server side:
http://some/url:port/connect

You will get "Hello" from server side:
{"jsonrpc":"2.0","method":"connected","params":{"time":"2018-09-06T14:40:28.06868112Z","channel":"tunnel-1","tunnel":"tunnel-1","text":"Hello!"}}

You can send request to create new exec:
{"jsonrpc":"2.0","id":0,"method":"create","params":{"identifier":{"machineName":"ws/machine-exec","workspaceId":"workspacen4b1ik9bxz1gqnoe"},"cmd":["/bin/bash"],"tty":true}}
You will get response(If it was successfully):
{"jsonrpc":"2.0","id":0,"result":1} // created exec with id: 1

resize exec:
request:
{"jsonrpc":"2.0","id":2,"method":"resize","params":{"id":2,"cols":235,"rows":24}}
response:
{"jsonrpc":"2.0","id":2,"result":{"id":2,"text":"Exec with id 2  was successfully resized"}}

check if exec is still alive (useful for restore output feature):

To attach and get output you should created separated websocket connection:


