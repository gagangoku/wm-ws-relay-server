package main

import (
	"flag"
	"log"

	"whatlist.io/whatsapp-proxy/server"
)

var listenPortFlag = flag.String("listenPort", "", "the port to listen on")
var externalEndpointFlag = flag.String("externalEndpoint", "", "the external endpoint for relay")

func main() {
	flag.Parse()

	log.Println("server starting: ", *listenPortFlag, *externalEndpointFlag)
	server.ServerMain(*listenPortFlag, *externalEndpointFlag)
}
