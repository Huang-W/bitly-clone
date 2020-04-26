##
## Docker Compose Deployment
##


up:
	docker-compose up -d

down:
	docker-compose down

network-ls:
	docker network ls

network-create:
	docker network create bitly

network-prune:
	docker network prune

#
# Docker Tag
#

tag-pod:
	docker tag wardhuang/datastore-worker:latest localhost:32000/datastore-worker

push-pod:
	docker push localhost:32000/datastore-worker

##
## Kubernetes Deployment
##

# Pod

pod-create:
	microk8s kubectl create -f cp-pod.yaml

pod-get:
	microk8s kubectl get pods --namespace bitly datastore-worker

pod-shell:
	microk8s kubectl exec --namespace bitly -it datastore-worker -- /bin/sh

pod-delete:
	microk8s kubectl delete pod --namespace bitly datastore-worker

# Deployment

deployment-create:
	microk8s kubectl create -f cp-deployment.yaml --save-config

deployment-get:
	microk8s kubectl get --namespace bitly deployments

deployment-get-pods:
	microk8s kubectl get pods --namespace bitly -l name=datastore-worker

deployment-pod-shell:
	microk8s kubectl exec  --namespace bitly -it $(pod) -- /bin/bash

deployment-upgrade:
	microk8s kubectl apply --namespace bitly -f cp-deployment.yaml

deployment-delete:
	microk8s kubectl delete deployment --namespace bitly cp-deployment

# Service

service-create:
	microk8s kubectl create -f cp-service.yaml

service-get:
	microk8s kubectl get --namespace bitly services

service-get-ip:
	microk8s kubectl get --namespace bitly service cp-service -o wide

service-delete:
	microk8s kubectl delete --namespace bitly service cp-service
