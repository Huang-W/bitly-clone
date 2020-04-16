
all: clean

clean:
	find . -name 'redirect-link' -type f -exec rm -f {} \;
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
	go get -v database/sql
	go get -v github.com/go-sql-driver/mysql

format:
	go fmt control-panel
	go fmt redirect-link

install:
	go install control-panel
	go install redirect-link

build:
	go build control-panel
	go build redirect-link

start-cp:
	./bin/control-panel

start-rl:
	./bin/redirect-link

docker-build-cp:
	docker build -t control-panel -f src/control-panel/Dockerfile .
	docker images

docker-build-rl:
	docker build -t redirect-link -f src/redirect-link/Dockerfile .

clean-api:
	docker stop control-panel
	docker stop redirect-link
	docker rm control-panel
	docker rm redirect-link

network-create:
	docker network create --driver bridge bitly

network-inspect:
	docker network inspect bitly

mysql-run:
	docker run --name mysql --network bitly -p 3306:3306 -e MYSQL_ROOT_PASSWORD=cmpe281 -td mysql:5.5

mysql-run-cp:
	docker run --name mysql-cp --network bitly -p 3307:3306 -e MYSQL_ROOT_PASSWORD=cmpe281 -td mysql:5.5

mongodb-run:
	docker run --name mongodb --network bitly -p 27017:27017 -td mongo

rabbitmq-run:
	docker run --name rabbitmq --network bitly --hostname my-rabbit \
						 -e RABBITMQ_DEFAULT_USER=user \
						 -e RABBITMQ_DEFAULT_PASS=password \
						 -p 8080:15672 -p 4369:4369 -p 5672:5672 \
						 -d rabbitmq:3-management


docker-run:
	docker run -d --name control-panel --network bitly -td -p 3000:3000 control-panel
	docker run -d --name redirect-link --network bitly -td -p 3001:3001 redirect-link

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

mysql-shell-cp:
	docker run -it --network bitly --rm mysql:5.5 mysql -h mysql-cp -u root -p

mongo-shell:
	docker run -it --rm --network bitly mongo \
						mongo --host mongodb \
						--authenticationDatabase admin \
						cmpe281

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
	 	localhost:3000/link_save \
		-H 'Content-Type: application/json' \
		-d '{"OrigUrl":"ifconfig.co"}'

test-get-origlink:
	curl -X GET \
	localhost:3001/r/1 \
	-H 'Content-Type: application/json'
