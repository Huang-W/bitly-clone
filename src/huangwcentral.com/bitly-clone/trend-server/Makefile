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
	docker tag wardhuang/trend-server:latest localhost:32000/trend-server

push-pod:
	docker push localhost:32000/trend-server

##
## Kubernetes Deployment
##

# Pod

pod-create:
	microk8s kubectl create -f ts-pod.yaml

pod-get:
	microk8s kubectl get pods --namespace bitly trend-server

pod-shell:
	microk8s kubectl exec --namespace bitly -it trend-server -- /bin/sh

pod-delete:
	microk8s kubectl delete pod --namespace bitly trend-server

# Deployment

deployment-create:
	microk8s kubectl create -f ts-deployment.yaml --save-config

deployment-get:
	microk8s kubectl get --namespace bitly deployments

deployment-get-pods:
	microk8s kubectl get pods --namespace bitly -l name=trend-server

deployment-pod-shell:
	microk8s kubectl exec  --namespace bitly -it $(pod) -- /bin/bash

deployment-upgrade:
	microk8s kubectl apply --namespace bitly -f ts-deployment.yaml

deployment-set:
	microk8s kubectl set image deployment ts-deployment trend-server=gcr.io/cmpe281-267121/trend-server

deployment-delete:
	microk8s kubectl delete deployment --namespace bitly ts-deployment

# Service

service-create:
	microk8s kubectl create -f ts-service.yaml

service-get:
	microk8s kubectl get --namespace bitly services

service-get-ip:
	microk8s kubectl get --namespace bitly service ts-service -o wide

service-delete:
	microk8s kubectl delete --namespace bitly service ts-service
