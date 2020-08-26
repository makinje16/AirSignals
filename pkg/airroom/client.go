package airroom

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

// AirMessageType is the kind of message being sent to a client
type AirMessageType string

const (
	// ClientMESSAGE is an enum type for an AirMessage
	ClientMESSAGE AirMessageType = "message"
	// ClientOFFER is an enum type for an AirMessage
	ClientOFFER AirMessageType = "offer"
	// ClientANSWER is an enum type for an AirMessage
	ClientANSWER AirMessageType = "answer"
	// ClientCANDIDATE is an enum type for an AirMessage
	ClientCANDIDATE AirMessageType = "candidate"
)

// AirClient is a representation of a single client that will attach to a room
type AirClient struct {
	//The unique identifier for the specific client
	ID string
	// the websocket connection to the client
	Conn *websocket.Conn
}

// AirMessage is the kind of struct allowed to be sent to a client
type AirMessage struct {
	MessageType AirMessageType
	Body        string
	SenderID    string
}

// NewAirMessage creates a message that is ok to send to an AirClient
func NewAirMessage(messageType AirMessageType, body string, senderID string) *AirMessage {
	return &AirMessage{
		MessageType: messageType,
		Body:        body,
		SenderID:    senderID,
	}
}

// NewEmptyAirMessage returns an AirMessage with no fields initialized
//
// Note: normally used for json Unmarshal
func NewEmptyAirMessage() *AirMessage {
	return &AirMessage{}
}

// NewClient creates a new client struct on the heap
func NewClient(id string, conn *websocket.Conn) *AirClient {
	var client *AirClient = new(AirClient)
	client.Conn = conn
	client.ID = id
	return client
}

// SendMessage sends a message to the client, c, over its websocket.Conn Conn
func (c *AirClient) SendMessage(message *AirMessage) {
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
	}
	c.Conn.WriteMessage(websocket.TextMessage, bytes)
}
