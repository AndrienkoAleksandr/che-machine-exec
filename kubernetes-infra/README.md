rm -rf vendor/ && rm -rf Godeps/ && godep save

godep restore ...

and `godep save` to project


docker build -t in-cluster2 .

kubectl run --rm -i exec --image=in-cluster2 --image-pull-policy=Never

kubectl delete deployment exec
