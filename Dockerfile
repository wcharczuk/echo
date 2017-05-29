FROM golang:1.8

ADD main.go /go/src/github.com/wcharczuk/echo/main.go
RUN go install github.com/wcharczuk/echo

WORKDIR "/go/src/github.com/wcharczuk/echo"
ENTRYPOINT /go/bin/echo
EXPOSE 5000