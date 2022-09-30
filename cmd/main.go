package main

import d7024e "D7024E_GRP9/kademlia"

//TODO Rename package structure.
//TODO FOR NODE LOOKUP -> PING RETURNING CANDIDATES

func main() {
	test_node := d7024e.NewKademlia("192.168.38.111", 8888)
	test_node.Run("192.168.38.201:8888", *d7024e.NewKademliaIDString("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"))

}
