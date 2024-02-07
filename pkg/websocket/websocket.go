package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all connections
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WebSocket struct {
	Upgrader websocket.Upgrader
	clients  map[*websocket.Conn]string
	lock     sync.Mutex
}

type Response struct {
	MessageType string   `json:"type"`
	Message     string   `json:"message"`
	FromUser    string   `json:"from_user"`
	UserList    []string `json:"online_users"`
}

func NewWebSocket() *WebSocket {
	return &WebSocket{
		Upgrader: upgrader,
		clients:  make(map[*websocket.Conn]string),
	}
}

// Takes in a custom responsewriter that has to have a hijack function
func (ws *WebSocket) HandleWebSocket(rw http.ResponseWriter, r *http.Request, user_id string) error {
	fmt.Println("WEBSOCKET INSIDE USER ID: ", user_id)
	if conn, err := upgrader.Upgrade(rw, r, nil); err != nil {
		return err
	} else {
		ws.lock.Lock()
		ws.clients[conn] = user_id
		ws.lock.Unlock()
		go ws.handleWebSocketConnection(conn)
	}

	return nil
}

func (ws *WebSocket) handleWebSocketConnection(conn *websocket.Conn) {
	defer func() {
		// ws.broadcastStatusUpdate("offline", ws.clients[conn])
		conn.Close()
		delete(ws.clients, conn)
	}()

	ws.lock.Lock()
	ws.broadcastStatusUpdate("online", ws.clients[conn])
	ws.lock.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("WebSocket Read Error:", err)
			ws.broadcastStatusUpdate("offline", ws.clients[conn])
			delete(ws.clients, conn)
			break
		}

		type ClientMessage struct {
			Message string `json:"message"`
			ToUser  string `json:"to_user"`
		}

		// Process the received message (e.g., broadcast to other users)
		var data ClientMessage
		err = json.Unmarshal(p, &data)
		if err != nil {
			fmt.Println("JSON ERROR:", err)
		}

		toUserConn := getKeyByValue(ws.clients, data.ToUser)
		if toUserConn == nil {
			fmt.Println("That user is not connected with WS")
		} else {
			jsonBytes, err := json.Marshal(Response{MessageType: "chat", Message: data.Message, FromUser: ws.clients[conn]})
			if err != nil {
				fmt.Println("JSON Marshal error!", err)
			}
			ws.sendToUser(messageType, []byte(jsonBytes), toUserConn)
		}
	}
}

func (ws *WebSocket) sendToUser(messType int, b []byte, userConn *websocket.Conn) {
	ws.lock.Lock()
	for conn := range ws.clients {
		if conn == userConn {
			go func(conn *websocket.Conn) {
				if err := conn.WriteMessage(messType, b); err != nil {
					fmt.Println("SOMETING ERROR", err)
				}
			}(conn)
		}
	}
	ws.lock.Unlock()
}

func (ws *WebSocket) broadcastStatusUpdate(status, username string) {
	ws.lock.Lock()
	var allConnections []string
	for w := range ws.clients {
		allConnections = append(allConnections, ws.clients[w])
	}
	ws.lock.Unlock()
	jsonBytes, err := json.Marshal(Response{MessageType: "status", Message: status, FromUser: username, UserList: allConnections})
	if err != nil {
		fmt.Println("JSON Marshal error!", err)
		return
	}
	ws.broadCastToAll(websocket.TextMessage, jsonBytes)
}

func (ws *WebSocket) broadCastToAll(messType int, b []byte) {
	ws.lock.Lock()
	for conn := range ws.clients {
		go func(conn *websocket.Conn) {
			if err := conn.WriteMessage(messType, b); err != nil {
				fmt.Println("SOMETING ERROR", err)
			}
		}(conn)
	}
	ws.lock.Unlock()
}

func getKeyByValue(m map[*websocket.Conn]string, targetValue string) *websocket.Conn {
	for key, value := range m {
		if value == targetValue {
			return key
		}
	}
	return nil
}