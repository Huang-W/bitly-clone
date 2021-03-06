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
	docker tag wardhuang/link-redirect:latest localhost:32000/link-redirect

push-pod:
	docker push localhost:32000/link-redirect

##
## Kubernetes Deployment
##

# Pod

pod-create:
	microk8s kubectl create -f lr-pod.yaml

pod-get:
	microk8s kubectl get pods --namespace bitly link-redirect

pod-shell:
	microk8s kubectl exec --namespace bitly -it link-redirect -- /bin/sh

pod-delete:
	microk8s kubectl delete pod --namespace bitly link-redirect

# Deployment

deployment-create:
	microk8s kubectl create -f lr-deployment.yaml --save-config

deployment-get:
	microk8s kubectl get --namespace bitly deployments

deployment-get-pods:
	microk8s kubectl get pods --namespace bitly -l name=link-redirect

deployment-pod-shell:
	microk8s kubectl exec  --namespace bitly -it $(pod) -- /bin/sh

deployment-upgrade:
	microk8s kubectl apply --namespace bitly -f lr-deployment.yaml

deployment-set:
	microk8s kubectl set image deployment lr-deployment link-redirect=gcr.io/cmpe281-267121/link-redirect

deployment-delete:
	microk8s kubectl delete deployment --namespace bitly lr-deployment

# Service

service-create:
	microk8s kubectl create -f lr-service.yaml

service-get:
	microk8s kubectl get --namespace bitly services

service-get-ip:
	microk8s kubectl get --namespace bitly service lr-service -o wide

service-delete:
	microk8s kubectl delete --namespace bitly service lr-service
