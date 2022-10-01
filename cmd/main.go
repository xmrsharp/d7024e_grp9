package main

import (
	d7024e "D7024E_GRP9/kademlia"
	"log"
	"net"
)

const (
	BOOT_LOADER_IP     = "172.19.0.2"
	BOOT_LOADER_STRING = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	PORT               = 8888
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

// NODE ID MATTERS ASWELL. SO NEED TO KNOW STATIC BOOTLOADER IP AND NODE ID.
// Check if i need to assign static kademlia id aswell
func main() {
	nodeIp := GetOutboundIP()
	BootLoaderId := d7024e.NewKademliaIDString(BOOT_LOADER_STRING)
	log.Println("IP:", nodeIp)
	if nodeIp.String() == BOOT_LOADER_IP {
		node := d7024e.NewKademlia(nodeIp.String(), PORT, BootLoaderId)
		node.Run("", *BootLoaderId)
	} else {
		node := d7024e.NewKademlia(nodeIp.String(), PORT, d7024e.NewRandomKademliaID())
		node.Run(BOOT_LOADER_IP+":8888", *BootLoaderId)
	}
	// test_node := d7024e.NewKademlia("192.168.38.111", 8888)
	// test_node.Run("192.168.38.201:8888", *d7024e.NewKademliaIDString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))

}
