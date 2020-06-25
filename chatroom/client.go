package chatroom

import (
	"github.com/gorilla/websocket"
)

// Client is a representation of a single client that will attach to a room
type Client struct {
	//The unique identifier for the specific client
	ID string
	// the websocket connection to the client
	Conn *websocket.Conn
}

// NewClient creates a new client struct on the heap
func NewClient(id string, conn *websocket.Conn) *Client {
	var client *Client = new(Client)
	client.Conn = conn
	client.ID = id
	return client
}
