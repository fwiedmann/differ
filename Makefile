clustername = differ-cluster
testVersion = 1.0.0
include .env
export
all: build buil_docker cluster_load_image cluster_deploy_testworkload run logs
dev: build run_dev
install_kind:
	curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/v0.6.1/kind-$$(uname)-amd64
	chmod +x ./kind
	sudo mv ./kind /usr/local/bin/kind

cluster_bootstrap: cluster_delete
	kind create cluster --config local-dev/local_cluster.yaml --wait 5m --name $(clustername)

cluster_delete:
	kind delete cluster --name differ-cluster

build:
	CGO_ENABLED=0 go build -o differ -ldflags "-X main.version=$(testVersion)"

buil_docker:
	docker build -t differ:dev .

cluster_load_image:
	kind load docker-image differ:dev --name $(clustername)

cluster_deploy_testworkload:
	kubectl delete -f local-dev/k8s; echo
	kubectl delete secret demo; echo
	kubectl create secret docker-registry demo \
	  --docker-server=$$DOCKER_REGISTRY_SERVER \
	  --docker-username=$$DOCKER_USER \
	  --docker-password=$$DOCKER_PASSWORD \
	  --docker-email=$$DOCKER_EMAIL
	kubectl apply -f local-dev/k8s/test-workload

run:
	kubectl apply -f local-dev/k8s
run_dev:
	./differ --devmode

logs:
	stern differ
