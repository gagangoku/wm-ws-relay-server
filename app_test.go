package main

import (
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"whatlist.io/whatsapp-proxy/client"
	"whatlist.io/whatsapp-proxy/server"
	"whatlist.io/whatsapp-proxy/util"
)

func Test_Test(t *testing.T) {
	t.Skip()
	hostPortFlag := "localhost:8080"
	ws1, _ := client.SimpleClient(fmt.Sprintf("ws://%s/echo", hostPortFlag), func(messageType int, data []byte) {
		fmt.Println("recv: ", string(data))
	})
	ws1.WriteMessage(websocket.TextMessage, []byte("test:hi"))

	time.Sleep(3 * time.Second)
	ws1.WriteMessage(websocket.TextMessage, []byte("test:hello"))
	time.Sleep(1 * time.Second)
}

func Test_Relay_ExistingUid(t *testing.T) {
	t.Parallel()
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("failed to get free port: %s", err)
	}
	hostPortFlag := fmt.Sprintf("localhost:%d", port)
	go func() {
		server.ServerMain(hostPortFlag)
	}()

	w1Recv, w2Recv := 0, 0
	ws1, _ := client.SimpleClient(fmt.Sprintf("ws://%s/registerUid?uid=123", hostPortFlag), func(messageType int, data []byte) {
		util.NoopFn(data)
		w1Recv++
	})
	ws2, _ := client.SimpleClient(fmt.Sprintf("ws://%s/relayExistingUid?uid=123", hostPortFlag), func(messageType int, data []byte) {
		util.NoopFn(data)
		w2Recv++
	})

	util.NoopFn(ws1, ws2)
	ws1.WriteMessage(websocket.TextMessage, []byte("hi"))

	time.Sleep(3 * time.Second)
	ws2.WriteMessage(websocket.TextMessage, []byte("hello"))

	for {
		if w1Recv > 0 && w2Recv > 0 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}

func Test_Relay_NewWebsocket(t *testing.T) {
	t.Parallel()
	port, err := GetFreePort()
	if err != nil {
		t.Fatalf("failed to get free port: %s", err)
	}
	hostPortFlag := fmt.Sprintf("localhost:%d", port)
	go func() {
		server.ServerMain(hostPortFlag)
	}()

	nRecv := 0
	u := fmt.Sprintf("ws://%s/relayNewWebsocket?wssUrl=%s", hostPortFlag, url.QueryEscape(fmt.Sprintf("ws://%s/echo", hostPortFlag)))
	ws, err := client.SimpleClient(u, func(messageType int, data []byte) {
		nRecv++
	})
	if err != nil {
		t.Fatalf("error in connecting to server: %s", err)
	}

	NUM := 10
	for i := 0; i < NUM; i++ {
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("iter %d", i)))
		time.Sleep(1 * time.Second)
	}
	if nRecv != NUM {
		t.Fatalf("something went wrong: %d", nRecv)
	}
}

// Source: https://gist.github.com/sevkin/96bdae9274465b2d09191384f86ef39d
func GetFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			port := l.Addr().(*net.TCPAddr).Port
			time.Sleep(100 * time.Millisecond)
			return port, nil
		}
	}
	return
}
