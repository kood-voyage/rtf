package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//Allow all connections
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocket struct {
	Upgrader websocket.Upgrader
	clients  map[*websocket.Conn]string
}

func NewWebSocket() *WebSocket {
	return &WebSocket{
		Upgrader: upgrader,
		clients:  make(map[*websocket.Conn]string),
	}
}

// Takes in a custom responsewriter that has to have a hijack function
func (ws *WebSocket) HandleWebSocket(rw http.ResponseWriter, r *http.Request, user_id string) error {
	fmt.Println(user_id)
	if conn, err := upgrader.Upgrade(rw, r, nil); err != nil {
		return err
	} else {
		ws.clients[conn] = user_id
		go ws.handleWebSocketConnection(conn)
	}

	return nil
}

func (ws *WebSocket) handleWebSocketConnection(conn *websocket.Conn) {
	defer func() {
		conn.Close()
		delete(ws.clients, conn)
	}()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket Read Error:", err)
			delete(ws.clients, conn)
			break
		}

		type ClientMessage struct {
			Message string `json:"message"`
			ToUser  string `json:"to_user"`
		}
		// fmt.Println(ws.clients)
		// Process the received message (e.g., broadcast to other users)
		var data ClientMessage
		err = json.Unmarshal(p, &data)
		if err != nil {
			fmt.Println("JSON ERROR:", err)
		}

		toUserConn := getKeyByValue(ws.clients, data.ToUser)
		if toUserConn == nil {
			fmt.Println("That user is not connected with WS")
		} else{
			ws.sendToUser(messageType, []byte(data.Message), toUserConn)
		}
	}
}

func (ws *WebSocket) sendToUser(messType int, b []byte, userConn *websocket.Conn) {
	for ws := range ws.clients {
		if (ws == userConn){
			go func(ws *websocket.Conn) {
				if err := ws.WriteMessage(messType, b); err != nil {
					fmt.Println("SOMETING ERROR", err)
				}
			}(ws)
		}
	}
}

func getKeyByValue(m map[*websocket.Conn]string, targetValue string) *websocket.Conn {
	for key, value := range m {
		if value == targetValue {
			return key
		}
	}
	return nil
}