package api

import (
	"log"
	"net/http"
	"strconv"
)

type APIServer struct {
	channelNode chan string
	server      *http.Server
}

func NewServer(addr string, port int) *APIServer {
	httpServer := http.Server{Addr: addr + ":" + strconv.Itoa(port)}
	handler := http.NewServeMux()
	httpServer.Handler = handler

	// Have to create instance first
	apiServer := APIServer{make(chan string), &httpServer}
	handler.HandleFunc("/test1", apiServer.endPointTest1)
	return &apiServer
}

func (as *APIServer) Run() {
	err := as.server.ListenAndServe()
	if err != nil {
		log.Println("COULDNT START API:", err)
	}

}

// TODO Endpoints:
// POST /objects - create object in http body
// BODY = VALUE
// RESPOND WITH LOCATION OF STORED OBJECT WITH JSON. 201 created. "Location: /objects/hash"
// GET /objects/{hash} - simply return with body of the key (basically hash is key).
// RESPOND WITH 200 - "hash: value"
func (as *APIServer) endPointTest1(w http.ResponseWriter, r *http.Request) {

	log.Println(r.Method)
	log.Println(r.Body)
	//response := http.Response{}

	w.Write([]byte("ENDPOINTTEST"))
}
