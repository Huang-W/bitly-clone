
all: clean

clean:
	find . -name 'link-redirect' -type f -exec rm -f {} \;
	find . -name 'control-panel' -type f -exec rm -f {} \;
	go clean

go-env:
	go env -w GOPATH=${CURDIR}

go-get:
	rm -rf src/github.com
	rm -rf src/gopkg.in
	go get -v github.com/codegangsta/negroni
	go get -v github.com/gorilla/mux
	go get -v github.com/unrolled/render
	go get -v github.com/satori/go.uuid
	go get -v github.com/gocql/gocql
	go get -v database/sql
	go get -v github.com/go-sql-driver/mysql

format:
	go fmt control-panel
	go fmt link-redirect
	go fmt trend-server
	go fmt datastore-worker

install:
	go install control-panel
	go install link-redirect
	go install trend-server
	go install datastore-worker

build:
	go build control-panel
	go build link-redirect
	go build trend-server
	go build datastore-worker

docker-build:
	docker build -t wardhuang/control-panel -f src/control-panel/Dockerfile .
	docker build -t wardhuang/link-redirect -f src/link-redirect/Dockerfile .
	docker build -t wardhuang/trend-server -f src/trend-server/Dockerfile .
	docker build -t wardhuang/datastore-worker -f src/datastore-worker/Dockerfile .
	docker images

docker-run:
	docker run -d --name control-panel --network bitly -td -p 3000:3000 wardhuang/control-panel
	docker run -d --name link-redirect --network bitly -td -p 3001:3001 wardhuang/link-redirect
	docker run -d --name trend-server --network bitly -td -p 3002:3002 wardhuang/trend-server
	docker run -d --name datastore-worker --network bitly -td wardhuang/datastore-worker

docker-clean:
	docker stop control-panel
	docker stop link-redirect
	docker stop trend-server
	docker stop datastore-worker
	docker rm control-panel
	docker rm link-redirect
	docker rm trend-server
	docker rm datastore-worker

log-cp:
	docker logs control-panel

log-lr:
	docker logs link-redirect

log-ts:
	docker logs trend-server

log-ds:
	docker logs datastore-worker

network-create:
	docker network create --driver bridge bitly

network-inspect:
	docker network inspect bitly

database-run:
	docker run --name nosql --network bitly -td -p 9090:9090 -p 8888:8888 wardhuang/nosql
	docker run --name mysql --network bitly -p 3306:3306 -e MYSQL_ROOT_PASSWORD=cmpe281 -td mysql:5.5
	docker run --name mongo-ts --network bitly -p 27017:27017 -td mongo
	docker run --name mongo-cp --network bitly -p 27019:27017 -td mongo
	docker run --name event-store --network bitly -p 27018:27017 -td mongo
	docker run --name rabbitmq --network bitly --hostname my-rabbit \
						 -e RABBITMQ_DEFAULT_USER=user \
						 -e RABBITMQ_DEFAULT_PASS=password \
						 -p 8080:15672 -p 4369:4369 -p 5672:5672 \
						 -d rabbitmq:3-management

kong-database:
	docker run -d --name kong-database --network bitly -p 9042:9042 cassandra:2.2

kong-run:
	docker run -d --name kong \
	          --network bitly \
              -e "KONG_DATABASE=cassandra" \
              -e "KONG_CASSANDRA_CONTACT_POINTS=kong-database" \
              -e "KONG_PG_HOST=kong-database" \
              -p 8000:8000 \
              -p 8443:8443 \
              -p 8001:8001 \
              -p 7946:7946 \
              -p 7946:7946/udp \
              kong:0.9.9

ts-shell:
	docker exec -it trend-server bash

cp-shell:
	docker exec -it control-panel bash

lr-shell:
	docker exec -it link-redirect bash

cloud-shell-ts:
	mongo -u cmpe281 --host 35.232.250.199

cloud-shell-cp:
	mongo -u cmpe281 --host 34.70.14.136

kong-shell:
	docker exec -it kong bash

mysql-shell:
	docker run -it --network bitly --rm mysql:5.5 mysql -h mysql -u root -p

mongo-shell-cloud:
	mongo "mongodb+srv://mongostorage-cscrb.mongodb.net/test"  --username administrator

mongo-shell-ts:
	docker run -it --rm --network bitly mongo \
						mongo --host mongo-ts \
						--authenticationDatabase admin \
						cmpe281

mongo-shell-cp:
	docker run -it --rm --network bitly mongo \
						mongo --host mongo-cp \
						--authenticationDatabase admin \
						cmpe281

mongo-shell-event:
	docker run -it --rm --network bitly mongo \
						mongo --host event-store \
						--authenticationDatabase admin \
						cmpe281

docker-network:
	docker network ls

docker-network-prune:
	docker network prune

docker-network-inspect:
	docker network inspect host

docker-ip:
	docker-machine ip

##
## API Test (Docker Compose / Kubernetes)
##

test-ping:
	curl localhost:3000/ping

test-create-shortlink:
	curl -X POST \
	 	localhost:3000/link_save \
		-H 'Content-Type: application/json' \
		-d '{"OrigUrl":"stackoverflow.com"}'

test-get-origlink:
	curl -X GET \
	localhost:3001/r/$(sl) \
	-H 'Content-Type: application/json'

test-get-trend:
	curl -X GET \
	localhost:3002/t/$(sl) \
	-H 'Content-Type: application/json'

##
## Kubernetes (Docker for Mac)
##

clean-up:
	microk8s kubectl delete --all pods --namespace=bitly
	microk8s kubectl delete --all deployments --namespace=bitly
	microk8s kubectl delete --all services --namespace=bitly

version:
	microk8s kubectl version

cluster:
	microk8s kubectl cluster-info

config:
	microk8s kubectl config view

nodes:
	microk8s kubectl get nodes

list-pods:
	microk8s kubectl get pods --namespace=bitly

list-all-pods:
	microk8s kubectl get pods --all-namespaces

list-system-pods:
	microk8s kubectl get pods --namespace=kube-system

install-dashboard:
	microk8s kubectl create -f kubernetes-dashboard.yaml

run-dashboard:
	microk8s kubectl port-forward $(pod) 8443:8443 --namespace=kube-system

start-api-proxy:
	microk8s kubectl proxy --port=8080

list-deployments:
	microk8s kubectl get deployments

describe-pod:
	microk8s kubectl describe pod $(pod)

create-namespace:
	microk8s kubectl create -f kubernetes-namespace.yaml

kube-namespace-services:
	microk8s kubectl get services -n bitly

docker-ps:
	 docker ps --all --format "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t"

docker-ps-ports:
	 docker ps --all --format "table {{.Names}}\t{{.Ports}}\t"



# Message Bus Pod

bus-create:
	microk8s kubectl create -f message-bus.yaml

bus-get:
	microk8s kubectl get --namespace bitly pod message-bus

bus-shell:
	microk8s kubectl exec  --namespace bitly -it message-bus -- /bin/sh

bus-delete:
	microk8s kubectl delete --namespace bitly pod message-bus

bus-docker:
	docker run --name message-bus -d gcr.io/cmpe281-267121/message-bus



##
## API Test inside Jump Box (Kubernetes Serivce)
##


jumpbox-ping:
	curl http://cp-service:9000/ping

jumpbox-create-shortlink:
	curl -X POST http://cp-service:9000/link_save -H 'Content-Type: application/json' -d '{"OrigUrl":"gobyexample.com"}'

jumpbox-get-shortlink:
	curl -X GET http://lr-service:9000/r/$(sl) -H 'Content-Type: application/json'

jumpbox-get-trend:
	curl -X GET http://ts-service:9000/t/$(sl) -H 'Content-Type: application/json'

jumpbox-order-status:
	curl -X GET \
  	http://gumball-service:9000/order \
  	-H 'Content-Type: application/json'

jumpbox-process-order:
	curl -X POST \
  	http://gumball-service:9000/orders \
  	-H 'Content-Type: application/json'
