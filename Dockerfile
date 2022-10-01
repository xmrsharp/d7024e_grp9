FROM golang:1.16-alpine

WORKDIR /app_kademlia


COPY go.mod ./
COPY go.sum ./

RUN go mod download

ADD kademlia ./kademlia
ADD cmd ./cmd

# RUN touch ./log_node
RUN go build -o /bin/kademlia_node ./cmd/main.go

EXPOSE 8888

CMD ["/bin/kademlia_node"]
# , "2>>/app_kademlia/log_node"]


