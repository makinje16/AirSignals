package airroom

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type clientMessageType string

const (
	ClientMESSAGE   clientMessageType = "message"
	ClientOFFER     clientMessageType = "offer"
	ClientANSWER    clientMessageType = "answer"
	ClientCANDIDATE clientMessageType = "candidate"
)

// AirClient is a representation of a single client that will attach to a room
type AirClient struct {
	//The unique identifier for the specific client
	ID string
	// the websocket connection to the client
	Conn *websocket.Conn
}

// ClientMessage is the kind of struct allowed to be sent to a client
type ClientMessage struct {
	messageType clientMessageType
	body        string
}

// NewClientMessage creates a message that is ok to send to an AirClient
func NewClientMessage(messageType clientMessageType, body string) ClientMessage {
	return ClientMessage{
		messageType: messageType,
		body:        body,
	}
}

// NewClient creates a new client struct on the heap
func NewClient(id string, conn *websocket.Conn) *AirClient {
	var client *AirClient = new(AirClient)
	client.Conn = conn
	client.ID = id
	return client
}

// SendMessage sends a message to the client, c, over its websocket.Conn Conn
func (c *AirClient) SendMessage(message ClientMessage) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	c.Conn.WriteMessage(websocket.TextMessage, bytes)
}
