package api

import (
	"encoding/json"
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
	// Have to create instance first
	apiServer := APIServer{ch, &httpServer}

	// Endpoints
	handler.HandleFunc("/objects", apiServer.postObject)
	handler.HandleFunc("/objects/", apiServer.getObject)
	return &apiServer
}

func (as *APIServer) Listen() {
	log.Println("API SERVING ON:", as.server.Addr)
	err := as.server.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
}

// TODO Fix correct encoding of payloads of /objects & /objects/{hash}
// TODO Fix status codes of response POST

type postObjectResponse struct {
	Location string `json:location`
}

type getObjectResponse struct {
	Value string `json:value`
}

// POST /objects - create object in http body
func (as *APIServer) postObject(w http.ResponseWriter, r *http.Request) {
	// value := r.Body
	if r.Method != "POST" {
		log.Println("DID NOT RECIEVE POST")
		return
	}
	nodeChannelMsg := APIChannel{ApiResponseChannel: make(chan string),
		ApiRequestMsg: "INCOMING POST REQUEST WITH BODY"}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	nodeResp := <-nodeChannelMsg.ApiResponseChannel

	log.Println("RECIEVED RESP:", nodeResp)
	w.Header().Set("Content-Type", "application/json")
	respBody := postObjectResponse{Location: "/objects/" + "THIS_IS_A_TEST_VALUE"}
	json.NewEncoder(w).Encode(respBody)
}

// GET /objects/{hash} - simply return with body of the key (basically hash is key).
// RESPOND WITH 200 - "hash: value"
// NOTE /objects/{hash} : GET value of hash if exist.
func (as *APIServer) getObject(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		log.Println("DID NOT RECIEVE GET")
		return
	}
	// NOTE This is will always work as /objects will be taken by other endpoint
	// That being said, do not try this at home
	temp := strings.Split(r.URL.String(), "/")
	hash := temp[2]

	nodeChannelMsg := APIChannel{ApiResponseChannel: make(chan string),
		ApiRequestMsg: "INCOMING GET REQUEST WITH HASH:" + hash}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	nodeResp := <-nodeChannelMsg.ApiResponseChannel

	log.Println("RECIEVED RESP:", nodeResp)

	w.Header().Set("Content-Type", "application/json")
	payload_test := getObjectResponse{Value: "THIS_IS_A_TEST_VALUE"}
	json.NewEncoder(w).Encode(payload_test)
}
