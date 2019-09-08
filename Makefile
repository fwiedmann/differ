all: cluster_bootstrap

install_kind:
	curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.5.1/kind-$$(uname)-amd64
	chmod +x ./kind
	sudo mv ./kind /usr/local/bin/kind

cluster_bootstrap:
	kind create cluster --config local_cluster.yaml --wait 5m --name differ-cluster

cluster_delete:
	kind delete cluster --name differ-cluster