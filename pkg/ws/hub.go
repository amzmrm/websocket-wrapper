package ws

import "sync"

var hub *Hub

// Hub manages all the registration and un-registration of WebSocket connections
type Hub struct {
	clients    sync.Map
	register   chan *Client
	unregister chan *Client
	handler    Handler
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients.Store(client.connID, client)
		case client := <-h.unregister:
			h.clients.Delete(client.connID)
		}
	}
}

func newHub(handler Handler) *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    sync.Map{},
		handler:    handler,
	}
}
