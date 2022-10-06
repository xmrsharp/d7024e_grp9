package main

import (
	"D7024E_GRP9/kademlia"
	"log"
	"net"
	"strconv"
)

// NOTE Assuming no other hosts currently on 172.20.0.0/16 network.
const (
	BOOT_LOADER_IP     = "172.20.0.2"
	BOOT_LOADER_STRING = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	PORT_KAD           = 8888
	PORT_API           = 8889
)

// Stolen from Stackoverflow.
func GetOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func main() {
	//nodeIp := GetOutboundIP()
	nodeIp := "127.0.0.1" // TODO When back to IP -> Insert .String() for checks.For testing purpouses

	BootLoaderId := kademlia.NewKademliaIDString(BOOT_LOADER_STRING)
	//nodeIp = nodeIp.String()
	if nodeIp == BOOT_LOADER_IP { // Insert .String() here at nodeIp.
		node := kademlia.NewKademlia(nodeIp, PORT_KAD, PORT_API, BootLoaderId) // Insert .String() here aswell
		node.Run("", *BootLoaderId)
	} else {
		node := kademlia.NewKademlia(nodeIp, PORT_KAD, PORT_API, kademlia.NewRandomKademliaID()) // Insert .String() here at nodeIP
		node.Run(BOOT_LOADER_IP+":"+strconv.Itoa(PORT_KAD), *BootLoaderId)
	}
}
