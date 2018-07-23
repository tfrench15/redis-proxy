FROM golang:onbuild
RUN redis-server
EXPOSE 8080
