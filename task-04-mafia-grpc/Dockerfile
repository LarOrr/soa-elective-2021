FROM golang:1.13 AS go-builder

WORKDIR /go/src/
COPY /src/. .

EXPOSE 8000

RUN go mod download
RUN go build -o ./app server.go

# TODO add FROM scratch

CMD ["./app"]
