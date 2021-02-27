package ws

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// RequestCode represents a WebSocket API
type RequestCode string

// WsRequest represents a WebSocket API and the parameters it requires
type Request struct {
	Code RequestCode     `json:"code"`
	Body json.RawMessage `json:"body"`
}

// Response is the response for a WebSocket request
type Response struct {
	Code    RequestCode `json:"code"`
	Error   bool        `json:"error"`
	Message *string     `json:"message,omitempty"`
	Data    interface{} `json:"content,omitempty"`
}

// NewWebSocketConn upgrades the connection from HTTP to WebSocket
func NewWebSocketConn(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rnd, err := uuid.NewRandom()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	c := &Client{
		hub:    hub,
		conn:   conn,
		send:   make(chan []byte, 1),
		connID: rnd.String()[:8],
	}
	c.hub.register <- c

	go c.writePump()
	go c.readPump()
}
