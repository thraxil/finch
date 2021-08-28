ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

finch: *.go
	go build .

run: finch .env
	. .env && ./finch

newdb:
	sqlite3 database.db < schema.sql

rmdb:
	rm -f database.db

build:
	docker run --rm -e CGO_ENABLED=true -e LDFLAGS='-extldflags "-static"' -v $(ROOT_DIR):/src -v /var/run/docker.sock:/var/run/docker.sock centurylink/golang-builder thraxil/finch

push: build
	docker push thraxil/finch
