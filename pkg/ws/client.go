package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var (
	writeWait      = time.Second * time.Duration(10)
	pongWait       = time.Second * time.Duration(10)
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = int64(1024 * 10)
	newline        = []byte{'\n'}
	space          = []byte{' '}
	upgrader       = websocket.Upgrader{
		ReadBufferSize:  1024 * 8,
		WriteBufferSize: 1024 * 8,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Client is a client that is connected with server via WebSocket
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	connID string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			log.Printf("failed to close WebSocket connection: %v", err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(
		func(string) error {
			c.conn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		},
	)

	var err error
	for {
		req := &Request{}
		err = c.conn.ReadJSON(req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				log.Printf("unexpected WebSocket error: %v", err)
			}
			break
		}
		c.hub.handler.ServeWs(c, req)
	}
}

func (c *Client) writePump() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic: %v", err)
			c.hub.unregister <- c
			c.conn.Close()
		}
	}()
	var err error
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err = w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err = c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// SendResponse marshal response and send
func (c *Client) SendResponse(resp *Response) {
	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("failed to marshal response: %v", err)
		return
	}
	c.send <- b
}
