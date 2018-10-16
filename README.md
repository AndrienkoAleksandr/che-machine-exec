# CHE machine exec

Go-lang server side to creation machine-execs for Eclipse CHE workspaces.
Uses to spawn terminal or command processes.

CHE machine exec uses json-rpc protocol to communication with client.

# How to use machine-exec image with Eclipse CHE workspace on the docker infrastructure:
Apply docker.sock path (by default it's `/var/run/docker.sock`) to the workspace volume property `CHE_WORKSPACE_VOLUME` in the che.env file:
Example:
 ```bash
CHE_WORKSPACE_VOLUME=/var/run/docker.sock:/var/run/docker.sock;
```
che.env file located in the CHE `data` folder. che.env file contains configuration properties for Eclipse CHE. All changes of the file become avaliable after restart Eclipse CHE.

Than run Eclipse CHE.
You can create new Eclipse CHE workspace with integreated Theia IDE from stack ''

By command pallette you can create new multimachine terminal: 
 
# Build docker image

Build docker image with che-machine-exec manually:

```bash
docker build -t eclipse/che-machine-exec .
```

# Run docker container

Run docker container with che-machine-exec manually:

```bash
docker run --rm -p 4444:4444 -v /var/run/docker.sock:/var/run/docker.sock eclipse/che-machine-exec
```

# Test che-machine-exec on the openshift locally
To test che-machine-exec you need deploy Eclipse CHE to the openshift locally. 
First of all prepare CHE to deploy. # todo link




To deploy Eclipse CHE to the local running openshift you can use [ocp.sh sript](https://github.com/eclipse/che/blob/master/deploy/openshift/ocp.sh). Run it with arguments:

```bash
./ocp.sh --run-ocp --deploy-che --no-pull --debug --deploy-che-plugin-registry --multiuser
```
In the output you will get link to the deployed Eclipse CHE project. Use it to login to Eclipse CHE. 
> Notice: for ocp.sh you could use argument `--setup-ocp-oauth`, but in this case you should use "Openshift v3" auth on the login page.

Register new user on the login page. After login you will be redirected to
the Eclipse CHE user dashboard. 

Create new workspace from openshift stack 'Java Theia on OpenShift' or (CHE 7 Preview) stack. Run workspace. When workspace will be running you will see Theia IDE. 

Create new terminal with help main menu: `File` => `New multy-machine terminal`. After that IDE propse for you select machine to creation terminal. Select one of the machine by click. After that Theia should create new terminal on the bottom panel.

Also you can create new Theia task for your project. In the project root create folder `.theia`. Create `tasks.json` file in the folder `.theia` with such content:

```bash
{
    "tasks": [
        {
            "label": "che",
            "type": "che",
            "command": "echo hello"
        }
    ]
}
```
and run it with help menu tasks: `Task` => `Run...`
After that Theia should display widget with output content: 'echo hello'

# Test on the Minishift

Install minishift with help this instractions:
 - https://docs.okd.io/latest/minishift/getting-started/preparing-to-install.html
 - https://docs.okd.io/latest/minishift/getting-started/setting-up-virtualization-environment.html


# todo prepare CHE to deploy

Install oc tool: [download oc binary for your platform](https://github.com/openshift/origin/releases), extract and apply this binary path to the system environment variables PATH. After that oc become availiable from terminal:

```bash
$ oc version
oc v3.9.0+191fece
kubernetes v1.9.1+a0ce1bc657
features: Basic-Auth GSSAPI Kerberos SPNEGO
```

Start Minishift:
```bash
$ minishift start --memory=8GB
-- Starting local OpenShift cluster using 'kvm' hypervisor...
...
   OpenShift server started.
   The server is accessible via web console at:
       https://192.168.99.128:8443

   You are logged in as:
       User:     developer
       Password: developer

   To login as administrator:
       oc login -u system:admin
```

From this command output You need:
 - Minishift master url. In this case it's `https://192.168.42.159:8443`. Let's call it 'CHE_INFRA_KUBERNETES_MASTER__URL'. We can store this variable in the terminal session to use it for next commands:

 ```bash
 export CHE_INFRA_KUBERNETES_MASTER__URL=https://192.168.42.162:8443
 ```
> Note: in case if you delete minishift virtual machine(`minishift delete`) and create it again, this url will be changed.

Register new user on the CHE_INFRA_KUBERNETES_MASTER__URL page.

Login to minishift with help oc, use new user login and password for it:

```bash
$ oc login --server=${CHE_INFRA_KUBERNETES_MASTER__URL}
```
This command activate openshift context to use minishift instance:

To deploy Eclipse CHE you can use [deploy.sh script](https://github.com/eclipse/che/blob/master/deploy/openshift/ocp.sh).

Run ocp.sh script with arguments:

```bash
export CHE_INFRA_KUBERNETES_MASTER__URL=${CHE_INFRA_KUBERNETES_MASTER__URL} && ./deploy_che.sh --no-pull --debug --multiuser
```

// Todo test ui

# Test on the Kubernetes (MiniKube)
Install minikube virtual machine on you computer: https://kubernetes-cn.github.io/docs/tasks/tools/install-minikube

You can install Eclipse CHE with help helm: https://github.com/eclipse/che/tree/master/deploy/kubernetes/helm/che#deploy-single-user-che-to-kubernetes-using-helm

So, Install helm: #Todo

Start new minikube:
```bash
minikube start --cpus 2 --memory 8192 --extra-config=apiserver.authorization-mode=RBAC
```

Go to helm/che directory:
```bash
$ cd ~/projects/che/deploy/kubernetes/helm/che
```

- Add cluster-admin role for `kube-system:default` account
```bash
kubectl create clusterrolebinding add-on-cluster-admin --clusterrole=cluster-admin --serviceaccount=kube-system:default
```
- Set your default Kubernetes context:
```bash
kubectl config use-context minikube
```
- Install tiller on your cluster:
  - Create a [tiller serviceAccount]: 
    ```bash
    kubectl create serviceaccount tiller --namespace kube-system
    ```
   - Bind it to the almighty cluster-admin role:
      ```bash
      kubectl apply -f ./tiller-rbac.yaml
      ```
  - Install tiller itself: 
    ```bash
    helm init --service-account tiller
    ```
- Start NGINX-based ingress controller:
  ```bash
  minikube addons enable ingress
  ```

There two configurations to deploy CHE on the Kubernetes:
 - first one: for eache new workspace Eclipse CHE creates separated namespace. In this case che-machine-exec doesn't work from the box yet, because: https://github.com/eclipse/che/issues/11349:
    ```bash
      helm upgrade --install che --namespace che --set global.ingressDomain=<domain> ./
    ```
 - second one: Eclipse CHE creates workspace in the same namespace. In this case che-machine-exec should work from the box. Deploy Eclipse CHE with command:
    ```bash
    helm upgrade --install che --namespace=che --set global.cheWorkspacesNamespace=che ./
    ```
> Notice you can track deploy CHE with help Minikube dashboard:
  ```bash
  minikube dashboard
  ```    

When Eclipse CHE will be deployed on the Minikube you can test ui:
for this stuff you should use kubernetes stack: 

# Test che-machine exec with help UI on the openshift/kubernetes inside Eclipse CHE.


 1208  helm delete che
 1209  helm delete eclipse-che
 1211  helm list
 1212  helm delete che6.13
 1213  helm list
 1214  cd projects/
 1215  cd che
 1216  cd deploy/kubernetes/helm/che/
 1217  kubectl create serviceaccount tiller --namespace kube-system
 1218  kubectl apply -f ./tiller-rbac.ya
 1219  kubectl apply -f ./tiller-rbac.yaml
 1220  helm init --service-account tiller
 1221  helm upgrade --install che6.13 --namespace eclips-che ./
 1222  helm upgrade --install che6 --namespace eclips-che ./
 1224  helm list
 1225  kubectl get pods
 1226  kubectl get pod
 1227  kubectl get pods
 1228  kubectl get pods 
 1229  kubectl get pods --namespace=all
 1230  kubectl get pods
 1231  kubectl get services
 1232  kubectl describe services/kubernetes
 1233  kubectl get pods -l
 1234  kubectl get pod -l
 1235  kubectl get pod -l app=eclipse-che
 1236  kubectl get pod -l app=v1
 1237  minikube get pods
 1238  minikube get pod
 1239  kubectl expose deployment eclipse-che --type=LoadBalancer
 1240  kubectl expose deployment eclipse-che--type=LoadBalancer
 1241  kubectl expose deployment che --type=LoadBalancer
 1242  kubectl expose deployment  --type=LoadBalancer
 1243  kubectl expose deployment che-dc7db84fb  --type=LoadBalancer
 1244  kubectl get pods
 1245  kubectl get pod
 1246  kubectl --help
 1247  kubectl get pods --help
 1248  kubectl get pods --all-namespaces
 1249  kubectl get deployment --all-namespaces
 1250  kubectl expose deployment che  --all-namespaces
 1251  kubectl expose deployment che --type=LoadBalancer
 1252  kubectl get deployment --all-namespaces
 1253  kubectl get ingresses --all-namespaces
 1254  kubectl enable addon ingress
 1255  minikube addons enable ingress
 1256  kubectl get pods --all-namespaces
 1257  kubectl get ingresses --all-namespaces
 1258  history

helm upgrade --install che --namespace=che --set global.cheWorkspacesNamespace=che ./


{
  "environments": {
    "default": {
      "machines": {
        "ws/theia": {
          "attributes": {},
          "servers": {
            "theia": {
              "protocol": "http",
              "port": "3000",
              "path": "/",
              "attributes": {
                "type": "ide"
              }
            },
            "theia-dev": {
              "protocol": "http",
              "port": "3030",
              "attributes": {
                "type": "ide-dev"
              }
            }
          },
          "volumes": {
            "projects": {
              "path": "/projects"
            }
          },
          "installers": [],
          "env": {
            "HOSTED_PLUGIN_HOSTNAME": "0.0.0.0"
          }
        },
        "ws/machine-exec": {
          "servers": {
            "machine-exec": {
              "attributes": {
                "type": "terminal"
              },
              "port": "4444",
              "protocol": "ws"
            }
          },
          "attributes": {}
        }
      },
      "recipe": {
        "type": "kubernetes",
        "content": "---\nkind: List\nitems:\n-\n  apiVersion: v1\n  kind: Pod\n  metadata:\n    name: ws\n  spec:\n    containers:\n      -\n        image: eclipse/che-theia:plugin-id-nightly\n        name: theia\n      -\n        image: wsskeleton/che-machine-exec\n        name: machine-exec\n",
        "contentType": "application/x-yaml"
      }
    }
  },
  "defaultEnv": "default",
  "name": "theia",
  "projects": [],
  "commands": []
}


# Prepare CHE to local deploy

> Requiements: install java 8 or higher and maven 3.3.0 or higher

First of all clone Eclipse che:

```
$ git clone https://github.com/eclipse/che.git ~/projects/che
```
For test purpose it's not nessure build all Eclipse CHE, assembly main is pretty enough:

```
$ cd assembly/assembly-main
$ mvn clean install -DskipTests
```