ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

finch: *.go
	go build .

install_deps:
	go get github.com/mattn/go-sqlite3
	go get github.com/gorilla/sessions
	go get github.com/russross/blackfriday
	go get github.com/nu7hatch/gouuid
	go get github.com/gorilla/feeds

deploy: finch
	rsync -avp media/ orlando.thraxil.org:/var/www/finch/media/
	rsync -avp templates/ orlando.thraxil.org:/var/www/finch/templates/
	rsync finch orlando.thraxil.org:/var/www/finch/finch
	rsync schema.sql orlando.thraxil.org:/var/www/finch/schema.sql
	ssh -t orlando.thraxil.org "sudo restart finch"

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
