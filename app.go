package main

import (
	"flag"
	"log"

	"whatlist.io/whatsapp-proxy/server"
)

var hostPortFlag = flag.String("hostPort", "", "the host port to listen to")
var protocolFlag = flag.String("protocol", "", "the protocol")

func main() {
	flag.Parse()

	log.Println("server starting: ", *hostPortFlag)
	server.ServerMain(*protocolFlag, *hostPortFlag)
}
