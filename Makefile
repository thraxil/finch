install_deps:
	go get github.com/mattn/go-sqlite3
	go get github.com/gorilla/sessions
	go get github.com/russross/blackfriday
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/feeds

run: finch .env
	. .env && ./finch

finch: *.go
	go build .

newdb:
	sqlite3 database.db < schema.sql

rmdb:
	rm -f database.db
