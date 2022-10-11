package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type APIChannel struct {
	ApiResponseChannel chan string
	ApiRequestMsg      string
}

type APIServer struct {
	channelNode chan<- APIChannel
	server      *http.Server
}

// Communication with node:
// Send APIChannelMSG trhough channelNode.
// And await reply in the Channel of the APIChannel
// That way we're lbocking on a specific channel.

func NewServer(addr string, port int, ch chan APIChannel) *APIServer {
	httpServer := http.Server{Addr: addr + ":" + strconv.Itoa(port)}
	handler := http.NewServeMux()
	httpServer.Handler = handler
	apiServer := APIServer{ch, &httpServer}

	// Endpoints
	handler.HandleFunc("/objects", apiServer.postObject)
	handler.HandleFunc("/objects/", apiServer.getObject)

	return &apiServer
}

func (as *APIServer) Listen() {
	log.Println("REST API SERVING ON:", as.server.Addr)
	err := as.server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

type getObjectResponse struct {
	Value string `json:value`
}

// POST /objects - create object in http body
func (as *APIServer) postObject(w http.ResponseWriter, r *http.Request) {
	// Read value to save.
	storeValue, _ := io.ReadAll(r.Body)

	if r.Method != "POST" || len(storeValue) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeChannelMsg := APIChannel{ApiResponseChannel: make(chan string),
		ApiRequestMsg: ""}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	nodeRsp := <-nodeChannelMsg.ApiResponseChannel
	// TODO Remove when making sure calls to create node and get node are working.
	log.Println("RECIEVED RESP:", nodeRsp)

	// Return response to caller
	w.Header().Add("Location", "/objects/"+nodeRsp)
	w.WriteHeader(http.StatusCreated)
	w.Write(storeValue)
}

// GET /objects/{hash} - simply return with body of the key (basically hash is key).
// RESPOND WITH 200 - "hash: value"
// NOTE /objects/{hash} : GET value of hash if exist.
func (as *APIServer) getObject(w http.ResponseWriter, r *http.Request) {
	// NOTE This is will always work as long as /objects will be taken by other endpoint
	// That being said, do not try this at home
	hashString := strings.Split(r.URL.String(), "/")[2]
	hash := []byte(hashString)

	// NOTE Hardcoded 20 byte value for byte id size.
	if r.Method != "GET" || len(hash) != 20 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	nodeChannelMsg := APIChannel{ApiResponseChannel: make(chan string),
		ApiRequestMsg: "INCOMING GET REQUEST WITH HASH:"}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	nodeRsp := <-nodeChannelMsg.ApiResponseChannel

	// TODO: Make sure storevalue and findnode are valid.
	log.Println("RECIEVED RESP:", nodeRsp)

	// Return response to caller.
	w.Header().Set("Content-Type", "application/json")
	body := getObjectResponse{Value: "THIS_IS_A_TEST_VALUE"}
	json.NewEncoder(w).Encode(body)

}
