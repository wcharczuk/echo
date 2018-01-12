FROM golang:1.8-alpine

RUN apk --no-cache add ca-certificates

WORKDIR "/go/src/github.com/wcharczuk/echo"

ADD main.go /go/src/github.com/wcharczuk/echo/main.go
RUN go get ./...
RUN go install github.com/wcharczuk/echo

ENTRYPOINT /go/bin/echo
EXPOSE 5000
