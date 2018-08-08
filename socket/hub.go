package socket

import (
	"encoding/json"
	"io"
)

var socketHub = newHub()

type hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client
}

type hubWriter struct {
	hub *hub
	tag string
}

func Writer(tag string) io.Writer {
	return &hubWriter{
		hub: socketHub,
		tag: tag,
	}
}

func (h *hubWriter) Write(p []byte) (n int, err error) {
	message := struct {
		Tag     string `json:"tag"`
		Message string `json:"bytes"`
	}{
		Tag:     h.tag,
		Message: string(p[:]),
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		return 0, err
	}

	return h.hub.Write(bytes)
}

func newHub() *hub {
	ret := &hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}

	go ret.start()
	return ret
}

func (h *hub) start() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		}
	}
}

func (h *hub) Write(p []byte) (n int, err error) {
	for client := range h.clients {
		select {
		case client.send <- p:
		default:
			delete(h.clients, client)
			close(client.send)
		}
	}

	return len(p), nil
}
