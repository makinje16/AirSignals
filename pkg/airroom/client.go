package airroom

import (
	"github.com/gorilla/websocket"
)

// AirClient is a representation of a single client that will attach to a room
type AirClient struct {
	//The unique identifier for the specific client
	ID string
	// the websocket connection to the client
	Conn *websocket.Conn
}

// NewClient creates a new client struct on the heap
func NewClient(id string, conn *websocket.Conn) *AirClient {
	var client *AirClient = new(AirClient)
	client.Conn = conn
	client.ID = id
	return client
}
