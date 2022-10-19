package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// TODO Upd request method similar to message to not hardcode values.
type APIChannel struct {
	ApiResponseChannel chan []byte
	//ApiRequestMethod   result.Result
	ApiRequestMethod  string
	ApiRequestPayload []byte
}

type APIServer struct {
	channelNode chan<- APIChannel
	server      *http.Server
}

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
	Value string `json:Value`
}

// POST /objects - create object in http body
// Return body of created object with location header of location.
func (as *APIServer) postObject(w http.ResponseWriter, r *http.Request) {
	// Read value to save.
	storeValue, _ := ioutil.ReadAll(r.Body)

	if r.Method != "POST" || len(storeValue) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	nodeChannelMsg := APIChannel{
		ApiResponseChannel: make(chan []byte),
		ApiRequestMethod:   "STORE_VALUE",
		ApiRequestPayload:  storeValue,
	}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	key := <-nodeChannelMsg.ApiResponseChannel
	if len(key) != 20 {
		log.Println("TODO MAKE SURE THIS WORKS WHEN CREATE NODE IMPLEMENTED.")
		// Failed to store value.
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO Remove when making sure calls to create node and get node are working.
	log.Println("RECIEVED RESP:", key)

	// Return response to caller
	w.Header().Add("Location", "/objects/"+string(key[:]))
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
	nodeChannelMsg := APIChannel{
		ApiResponseChannel: make(chan []byte),
		ApiRequestMethod:   "GET_VALUE",
		ApiRequestPayload:  hash,
	}

	// Write msg to node.
	as.channelNode <- nodeChannelMsg
	// Await response in wrapped channel
	keyValue := <-nodeChannelMsg.ApiResponseChannel
	if len(keyValue) == 0 {
		// Failed to get hash value.
		w.WriteHeader(http.StatusBadRequest)
	}
	// TODO: Make sure storevalue and findnode are valid.
	log.Println("RECIEVED RESP:", keyValue)

	// Return response to caller.
	w.Header().Set("Content-Type", "application/json")
	// TODO Replace when node functions tested and done.
	body := getObjectResponse{Value: "TODO REPLACE WHEN FIND VALUE IS IMPLEMENTED IN NODE."}
	json.NewEncoder(w).Encode(body)

}
