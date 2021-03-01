FROM golang:1.13

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN go build -o ./app ser_bench

CMD ["./app"]