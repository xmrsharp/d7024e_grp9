package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	KEY_STRING_LENGTH = 40
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

func parseNodeResponse(resp []byte) (string, string) {
	respString := strings.Split(string(resp), " - ")
	return respString[0], respString[1]
}

// POST /objects - create object in http body
// Return body of created object with location header of location.
func (as *APIServer) postObject(w http.ResponseWriter, r *http.Request) {
	log.Println("RECIEVED POST REQUEST")
	// Read value to save.
	storeValue, _ := ioutil.ReadAll(r.Body)
	log.Println("STOREVALUE:", storeValue)
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
	nodeResp := <-nodeChannelMsg.ApiResponseChannel
	key, _ := parseNodeResponse(nodeResp)
	if len(key) != KEY_STRING_LENGTH {
		// Somehow got a bad response from node network.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return response to caller
	w.Header().Add("Location", "/objects/"+key)
	w.WriteHeader(http.StatusCreated)
	w.Write(storeValue)
}

// GET /objects/{hash} - simply return with body of the key (basically hash is key).
// RESPOND WITH 200 - "hash: value"
// NOTE /objects/{hash} : GET value of hash if exist.
func (as *APIServer) getObject(w http.ResponseWriter, r *http.Request) {
	log.Println("RECIEVED GET REQUEST")
	// NOTE This is will always work as long as /objects will be taken by other endpoint
	// That being said, do not try this at home
	hashString := strings.Split(r.URL.String(), "/")[2]
	hash := []byte(hashString)
	// NOTE Hardcoded 20 byte value for byte id size.
	if r.Method != "GET" || len(hash) != KEY_STRING_LENGTH {
		log.Println("BAD REQUEST")
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
