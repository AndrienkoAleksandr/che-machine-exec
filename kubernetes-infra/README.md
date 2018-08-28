rm -rf vendor/ && rm -rf Godeps/ && godep save

godep restore ...

and `godep save` to project

eval $(minikube docker-env)

docker build -t in-cluster2 .

kubectl run --rm -i exec --image=in-cluster2 --port=4000 --image-pull-policy=Never


kubectl expose deployment exec --type=LoadBalancer --name=exec-service
# open in the browser service
minikube service exec-service

# apply /api/create

kubectl delete deployment exec
