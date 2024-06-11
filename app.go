package main

import (
	"flag"
	"log"

	"whatlist.io/whatsapp-proxy/server"
)

var hostPortFlag = flag.String("hostPort", "", "the host port to listen to")

func main() {
	log.Println("server starting: ", *hostPortFlag)
	server.ServerMain(*hostPortFlag)
}
