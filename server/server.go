package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"whatlist.io/whatsapp-proxy/client"

	"github.com/gorilla/websocket"
)

var injectorCode string
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var sockets = map[string]*websocket.Conn{}

func (app *App) echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		log.Printf("recv: %s, type: %s", message, websocket.FormatMessageType(mt))
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (app *App) relayNewWebsocket(w http.ResponseWriter, r *http.Request) {
	ws1, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %s\n", err)
		fmt.Fprintf(w, "upgrade error: %s", err)
		return
	}
	defer ws1.Close()
	wssUrl := r.URL.Query().Get("wssUrl")
	if wssUrl == "" {
		log.Print("wssUrl is must")
		fmt.Fprintf(w, "wssUrl is must")
		return
	}

	onMsgFn := func(messageType int, data []byte) {
		err := ws1.WriteMessage(messageType, data)
		if err != nil {
			log.Printf("failed to write to ws1: %s\n", err)
			fmt.Fprintf(w, "failed to write to ws1: %s", err)
			ws1.Close()
		}
	}
	ws2, err := client.SimpleClient(wssUrl, onMsgFn)
	if err != nil {
		log.Printf("failed to connect to %s: %s\n", wssUrl, err)
		fmt.Fprintf(w, "failed to connect to %s: %s\n", wssUrl, err)
		return
	}

	for {
		mt, message, err := ws1.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s, type: %s", message, websocket.FormatMessageType(mt))

		err = ws2.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (app *App) registerUid(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		log.Print("uid is must")
		serializeBaseRsp(false, fmt.Sprintf("uid is must"), "")
		return
	}

	ws1, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %s\n", err)
		serializeBaseRsp(false, fmt.Sprintf("upgrade error: %s", err), "")
		return
	}
	log.Printf("upgraded %p: registerUid=%s", ws1, uid)

	sockets[uid] = ws1
}

func (app *App) relayExistingUid(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		log.Print("uid is must")
		serializeBaseRsp(false, fmt.Sprintf("uid is must"), "")
		return
	}

	ws2, ok := sockets[uid]
	if !ok {
		log.Print("no websocket with uid")
		serializeBaseRsp(false, fmt.Sprintf("no websocket with uid"), "")
		return
	}

	ws1, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrade error: %s\n", err)
		serializeBaseRsp(false, fmt.Sprintf("upgrade error: %s", err), "")
		return
	}
	log.Printf("upgraded %p: relayExistingUid=%s", ws1, uid)

	go func() {
		pipe(ws1, ws2)
	}()
	pipe(ws2, ws1)
}

func pipe(ws1, ws2 *websocket.Conn) error {
	defer ws1.Close()
	defer ws2.Close()

	for {
		mt, message, err := ws1.ReadMessage()
		if err != nil {
			log.Printf("read error on %p: %s\n", ws1, err)
			return err
		}
		log.Printf("recv on %p: %s, type: %s", ws1, message, websocket.FormatMessageType(mt))

		err = ws2.WriteMessage(mt, message)
		if err != nil {
			log.Printf("write error on %p: %s\n", ws2, err)
			return err
		}
	}
}

func (app *App) injectWebsocketRelayInBrowser(w http.ResponseWriter, r *http.Request) {
	uid := r.URL.Query().Get("uid")
	if uid == "" {
		log.Print("uid is must")
		sendResponse(w, http.StatusBadRequest, serializeBaseRsp(false, fmt.Sprintf("uid is must"), ""))
		return
	}

	w.Header().Set("X-Server", "whatlist-websocket-relay")
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET,POST,OPTIONS")
	w.WriteHeader(http.StatusOK)

	s := strings.ReplaceAll(injectorCode, "__uid__", uid)
	s = strings.ReplaceAll(s, "__protocolHostport__", app.externalEndpoint)
	io.WriteString(w, s)
}

func home(w http.ResponseWriter, r *http.Request) {
	sendResponse(w, http.StatusOK, "hi")
}

func ServerMain(listenPort, externalEndpoint string) {
	log.Printf("starting server: %s %s\n", listenPort, externalEndpoint)

	_bytes, err := os.ReadFile("injector.html")
	if err != nil {
		log.Fatalf("Error: couldnt read injector code: %s", err)
		return
	}
	injectorCode = string(_bytes)

	app := &App{listenPort: listenPort, externalEndpoint: externalEndpoint}
	mux := http.NewServeMux()
	mux.HandleFunc("/echo", app.echo)
	mux.HandleFunc("/relayNewWebsocket", app.relayNewWebsocket)
	mux.HandleFunc("/relayExistingUid", app.relayExistingUid)
	mux.HandleFunc("/registerUid", app.registerUid)
	mux.HandleFunc("/injectWebsocketRelayInBrowser", app.injectWebsocketRelayInBrowser)
	mux.HandleFunc("/", home)

	ctx, cancelCtx := context.WithCancel(context.Background())
	serverOne := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%s", listenPort),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, ctxKey, l.Addr().String())
			return ctx
		},
	}

	err = serverOne.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		log.Println("Error: server one closed")
	} else if err != nil {
		log.Printf("Error listening for server: %s\n", err)
	}
	cancelCtx()
}

func sendResponse(w http.ResponseWriter, code int, output string) {
	w.Header().Set("X-Server", "whatlist-websocket-relay")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Method", "GET,POST,OPTIONS")
	w.WriteHeader(code)
	io.WriteString(w, output)
}

func serializeBaseRsp(success bool, errorMsg, rspStr string) string {
	rsp := BaseResponse{Success: success, ErrorMsg: errorMsg, Rsp: rspStr}
	bytes, _ := json.Marshal(rsp)
	return string(bytes)
}

type App struct {
	listenPort       string
	externalEndpoint string
}
type BaseResponse struct {
	Success  bool   `json:"success"`
	ErrorMsg string `json:"errorMsg,omitempty"`
	Rsp      string `json:"rsp,omitempty"`
}

type key int

const ctxKey key = iota
