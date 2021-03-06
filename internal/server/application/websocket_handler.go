package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	WS_REQ_GETSYNCHLIST = "getSynchsList"
	WS_REQ_STARTSYNCH   = "startSynch"
	WS_REQ_STOPSYNCH    = "stopSynch"
)

type wsInbound struct {
	ID   string        `yaml:"id"`
	Name string        `yaml:"name"`
	Data wsInboundData `yaml:"data"`
}

type wsInboundData struct {
	Payload string `yaml:"payload"`
}

type wsOutbound struct {
	ID      string         `yaml:"id"`
	Name    string         `yaml:"name"`
	Success bool           `yaml:"success"`
	Data    wsOutboundData `yaml:"data"`
}

type wsOutboundData struct {
	Message string `yaml:"message"`
	Payload string `yaml:"payload"`
}

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type webSocketHandler struct {
	app *Application
}

func (wsh *webSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsUpgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		panic(err)
	}
	log.Println("Client Connected")

	if wsh.app == nil {
		log.Println("err")
	}

	go wsh.wsReader(ws)
}

func (wsh *webSocketHandler) wsReader(ws *websocket.Conn) {
	for {
		messageType, message, err := ws.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		var wsReq wsInbound
		marshalErr := json.Unmarshal(message, &wsReq)
		if marshalErr != nil {
			// panic(marshalErr)
		}
		fmt.Println(wsReq)
		wsh.dispatchWsRequest(ws, &wsReq, messageType)
	}
}

func (wsh *webSocketHandler) dispatchWsRequest(ws *websocket.Conn, wsReq *wsInbound, messageType int) {
	var wsOut wsOutbound

	switch wsReq.Name {
	case WS_REQ_GETSYNCHLIST:
		synchsList := wsh.app.listSynchsToJSON()
		wsOut = wsOutbound{
			ID:      wsReq.ID,
			Name:    "synchsListFetched",
			Success: true,
			Data: wsOutboundData{
				Payload: string(synchsList),
			},
		}
	case WS_REQ_STARTSYNCH:
		synchName := wsReq.Data.Payload

		responseChan := createResponseChannel()
		go wsh.app.runSynch(responseChan, "one-off", synchName, true)
		// response := <-responseChan

		wsOut = wsOutbound{
			ID:      wsReq.ID,
			Name:    "synchStarted",
			Success: true,
			Data:    wsOutboundData{
				// Message: response.(string),
			},
		}
	case WS_REQ_STOPSYNCH:
		synchName := wsReq.Data.Payload

		responseChan := createResponseChannel()
		go wsh.app.stopSynch(responseChan, synchName)
		// response := <-responseChan

		wsOut = wsOutbound{
			ID:      wsReq.ID,
			Name:    "synchStopped",
			Success: true,
			Data:    wsOutboundData{
				// Message: response.(string),
			},
		}
	default:
		wsOut = wsOutbound{
			ID:      wsReq.ID,
			Name:    "unknownRequest",
			Success: false,
			Data: wsOutboundData{
				Message: "Unknown websocket request name \"" + wsReq.Name + "\".",
			},
		}
	}

	wsOutJSON, err := json.Marshal(wsOut)
	if err != nil {
		panic(err)
	}

	if err := ws.WriteMessage(messageType, wsOutJSON); err != nil {
		log.Println(err)
		return
	}
}
