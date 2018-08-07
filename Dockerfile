FROM golang

ADD . /go/src/github.com/tfrench15/redis-proxy

RUN go install github.com/tfrench15/redis-proxy

CMD ["make test"]

EXPOSE 8080
