install_deps:
	go get github.com/mattn/go-sqlite3
	go get github.com/gorilla/sessions
	go get github.com/russross/blackfriday

run: finch .env
	. .env && ./finch

finch: *.go
	go build .

newdb:
	sqlite3 database.db < schema.sql

rmdb:
	rm -f database.db
