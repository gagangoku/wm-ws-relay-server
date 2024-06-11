package client

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

func SimpleClient(remoteUrl string, onMsgFn func(messageType int, data []byte)) (*websocket.Conn, error) {
	parsedUrl, err := url.Parse(remoteUrl)
	if err != nil {
		return nil, err
	}
	log.Printf("connecting to %s", parsedUrl.String())

	c, _, err := websocket.DefaultDialer.Dial(parsedUrl.String(), nil)
	if err != nil {
		return nil, err
	}

	go func() {
		defer c.Close()
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read error: ", err)
				return
			}
			log.Printf("recv on %p: %s, type: %s", c, message, websocket.FormatMessageType(mt))
			onMsgFn(mt, message)
		}
	}()
	log.Printf("connected %p: %s", c, remoteUrl)
	return c, nil
}
