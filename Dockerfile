FROM golang:1.8

WORKDIR "/go/src/github.com/wcharczuk/echo"

ADD main.go /go/src/github.com/wcharczuk/echo/main.go
ADD vendor /go/src/github.com/wcharczuk/echo/vendor
RUN go build -o /go/bin/echo .

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=0 /go/bin/echo .
CMD [ "./echo" ]
EXPOSE 5000