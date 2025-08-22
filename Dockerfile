FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o kvstore .

EXPOSE 2222 8222 7469

ENTRYPOINT ["./kvstore"]
CMD ["--node-id", "node1", "--http-port", "2222", "--raft-port", "8222", "--gossip-port", "7469"]