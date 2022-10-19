package kademlia

import (
	"fmt"

	"testing"
)

const (
	BOOT_LOADER_IP     = "172.20.0.2"
	BOOT_LOADER_STRING = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	PORT_KAD           = 8888
	PORT_API           = 8889
)

func TestNewKademlia(t *testing.T) {
	fmt.Printf("Testingnewkademlia")
	nodeIp := "127.0.0.1"
	BootLoaderId := NewKademliaIDString(BOOT_LOADER_STRING)
	node := NewKademlia(nodeIp, PORT_KAD, PORT_API, BootLoaderId)
	fmt.Printf("Node: %T\n", node)

}
