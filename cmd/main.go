package main

import (
	"D7024E_GRP9/network"
)

func main() {
	server_test := network.Network{}
	server_test.Listen("127.0.0.1", 8888)
}
