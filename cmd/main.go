package main

import (
	d7024e "D7024E_GRP9/kademlia"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

// NOTE Assuming no other hosts currently on 172.20.0.0/16 network.
const (
	BOOT_LOADER_IP     = "172.20.0.2"
	BOOT_LOADER_STRING = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	PORT               = 8888
)

var out io.Writer = os.Stdout

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
	nodeIp := GetOutboundIP()
	BootLoaderId := d7024e.NewKademliaIDString(BOOT_LOADER_STRING)
	if nodeIp.String() == BOOT_LOADER_IP {
		node := d7024e.NewKademlia(nodeIp.String(), PORT, BootLoaderId)
		node.Run("", *BootLoaderId)
	} else {
		node := d7024e.NewKademlia(nodeIp.String(), PORT, d7024e.NewRandomKademliaID())
		node.Run(BOOT_LOADER_IP+":"+strconv.Itoa(PORT), *BootLoaderId)
	}
}
