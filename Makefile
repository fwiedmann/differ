clustername = differ-cluster

all: cluster_deploy

install_kind: 
	curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$$(uname)-amd64
	chmod +x ./kind
	sudo mv ./kind /usr/local/bin/kind

cluster_bootstrap:
	kind create cluster --config local-dev/local_cluster.yaml --wait 5m --name $(clustername)

cluster_delete:
	kind delete cluster --name differ-cluster

build: 
	CGO_ENABLED=0 go build -o differ
	docker build -t differ:dev .

cluster_load_image:
	kind load docker-image differ:dev --name $(clustername)

cluster_deploy: build cluster_load_image
	KUBECONFIG=$(shell kind get kubeconfig-path --name="differ-cluster")
	kubectl delete -f local-dev/k8s
	kubectl apply -f local-dev/k8s
