package kademlia

import (
	"fmt"
	"testing"
	"time"
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

func TestNodeLookup(t *testing.T) {
	fmt.Printf("Test nodelookup")
	nodeIp := "127.0.0.1"
	BootLoaderId := NewKademliaIDString(BOOT_LOADER_STRING)
	node := NewKademlia(nodeIp, PORT_KAD, PORT_API, BootLoaderId)

	contact := NewContact(NewKademliaIDString("FFFFFFFF00000000000000000000000000000000"), "localhost:8000")
	//rt := NewRoutingTable(contact)
	<-time.After(2 * time.Second)
	/*Create kademlia network for bootstrap node*/
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000000000000000000F"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000000000000000000F0"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000000000000000000F00"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000000000000000F000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000000000000000F0000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000000000000000F00000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000000000000F000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000000000000F0000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000000000000F00000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000000000F000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000000000F0000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000000000F00000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000000F000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000000F0000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000000F00000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000000F000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000000F0000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000000F00000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000000F000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000000F0000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000000F00000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000000F000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000000F0000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000000F00000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0000000F000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000000F0000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00000F00000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF000F0000000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF00F00000000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFF0F000000000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(NewContact(NewKademliaIDString("FFFFFFFFF0000000000000000000000000000000"), "localhost:8080"))
	node.routingTable.AddContact(contact)

	node.NodeLookup(contact.ID)
	node.genCheckBuckets()

	node.LookupData(*NewRandomKademliaID())

	val := "testarstore"
	key := NewKademliaID(&val)
	node.StoreValue(val)
	node.LookupData(key)

}
