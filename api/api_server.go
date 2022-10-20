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
	KEY_STRING_LENGTH  = 40
	ERROR_STRING_VALUE = "ERROR"
)

// TODO Upd request method similar to message to not hardcode values.
type APIChannel struct {
	ApiResponseChannel chan []byte
	ApiRequestMethod   string
	ApiRequestPayload  []byte
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
	Key   string `json:Key`
}

func parseNodeResponse(resp []byte) (string, string) {
	respString := string(resp)
	if respString == ERROR_STRING_VALUE {
		return ERROR_STRING_VALUE, ERROR_STRING_VALUE
	}
	respValues := strings.Split(respString, " - ")
	return respValues[0], respValues[1]
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
	nodeRsp := <-nodeChannelMsg.ApiResponseChannel
	key, _ := parseNodeResponse(nodeRsp)
	if len(key) != KEY_STRING_LENGTH || key == ERROR_STRING_VALUE {
		// Somehow got a bad response from node network.
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Return response to caller
	w.Header().Add("Location", "/objects/"+key)
	w.WriteHeader(http.StatusCreated)
	w.Write(storeValue)
}

// RESPOND WITH 200 - "hash: value"
// NOTE /objects/{hash} : GET value of hash if exist.
func (as *APIServer) getObject(w http.ResponseWriter, r *http.Request) {

	hashString := strings.Split(r.URL.String(), "/")[2]
	hash := []byte(hashString)

	// NOTE Hardcoded 20 byte value for byte id size.
	if r.Method != "GET" || len(hash) != KEY_STRING_LENGTH {
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
	nodeRsp := <-nodeChannelMsg.ApiResponseChannel
	key, val := parseNodeResponse(nodeRsp)
	if len(key) != KEY_STRING_LENGTH || key == ERROR_STRING_VALUE {
		// Failed to get hash value.
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO: Make sure storevalue and findnode are valid.
	log.Println("RECIEVED RESP:", key, val)

	// Return response to caller.
	w.Header().Set("Content-Type", "application/json")
	// TODO Replace when node functions tested and done.
	body := getObjectResponse{Key: key, Value: val}
	json.NewEncoder(w).Encode(body)

}
