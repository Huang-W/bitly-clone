
all: clean

clean:
	find . -name 'gumball' -type f -exec rm -f {} \;
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
	go get -v database/sql
	go get -v github.com/go-sql-driver/mysql

run:
	go run src/app/$(app).go

main:
	go run src/app/main.go

format:
	go fmt gumball

install:
	go install control-panel

build:
	go build control-panel

start:
	./gumball

test-ping:
	curl localhost:3000/ping

test-gumball:
	curl localhost:3000/gumball

docker-build:
	docker build -t control-panel -f src/control-panel/Dockerfile .
	docker images

network-create:
	docker network create --driver bridge bitly

network-inspect:
	docker network inspect bitly

mysql-run:
	docker run -d --name mysql --network bitly -td -p 3306:3306 -e MYSQL_ROOT_PASSWORD=cmpe281 mysql:5.5

rabbitmq-run:
	docker run -d --name rabbit --network bitly --hostname my-rabbit \
						 -e RABBITMQ_DEFAULT_USER=user \
						 -e RABBITMQ_DEFAULT_PASS=password \
						 -p 8080:15672 \
						 rabbitmq:3-management


docker-run:
	docker run -d --name control-panel --network bitly -td -p 3000:3000 control-panel

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

docker-shell:
	docker exec -it control-panel bash

kong-shell:
	docker exec -it kong bash

mysql-shell:
	docker run -it --network bitly --rm mysql:5.5 mysql -h mysql -u root -p

docker-network:
	docker network ls

docker-network-prune:
	docker network prune

docker-network-inspect:
	docker network inspect host

docker-clean:
	docker stop mysql
	docker rm mysql
	docker stop kong-database
	docker rm kong-database
	docker stop kong
	docker rm kong
	docker stop control-panel
	docker rm control-panel
	docker rmi control-panel

docker-ip:
	docker-machine ip

docker-ps:
	 docker ps --all --format "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.Status}}\t"

docker-ps-ports:
	 docker ps --all --format "table {{.Names}}\t{{.Ports}}\t"

test-create-shortlink:
	curl -X POST \
	 	localhost:3000/urls \
		-H 'Content-Type: application/json' \
		-d '{"OrigUrl":"ifconfig.co"}'

test-get-origlink:
	curl -X GET \
	localhost:3001/r/1 \
	-H 'Content-Type: application/json'
