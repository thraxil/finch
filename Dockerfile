FROM golang

RUN go get github.com/mattn/go-sqlite3 && \
    go get github.com/gorilla/sessions && \
    go get github.com/russross/blackfriday && \
    go get github.com/nu7hatch/gouuid && \
    go get github.com/gorilla/feeds && \
    go get code.google.com/p/go.crypto/bcrypt

ADD . /go/src/github.com/thraxil/finch
RUN go install github.com/thraxil/finch
RUN mkdir -p /var/lib/finch
ENV FINCH_PORT 8000
ENV FINCH_DB_FILE /var/lib/finch/database.db
ENV FINCH_MEDIA_DIR /go/src/github.com/thraxil/finch/media
ENV FINCH_ITEMS_PER_PAGE 50
VOLUME /var/lib/finch
EXPOSE 8000
CMD ["/go/bin/finch"]
