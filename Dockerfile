FROM golang:1.16-alpine

WORKDIR /app_kademlia


COPY go.mod ./
COPY go.sum ./

RUN go mod download

ADD kademlia ./kademlia
ADD cmd ./cmd

RUN go build -o /bin/kademlia_node ./cmd/main.go

EXPOSE 8888

CMD ["/bin/kademlia_node"]