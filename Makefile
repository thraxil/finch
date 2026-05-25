ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

finch: *.go
	go build .

test:
	go test -v ./...

run: finch
	./finch

newdb:
	sqlite3 database.db < schema.sql

seed:
	sqlite3 database.db < seed.sql

rmdb:
	rm -f database.db

deploy:
	~/.fly/bin/flyctl deploy
