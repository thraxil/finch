ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

finch: *.go
	go build .

run: finch .env
	. .env && ./finch

newdb:
	sqlite3 database.db < schema.sql

rmdb:
	rm -f database.db

deploy:
	~/.fly/bin/flyctl deploy
