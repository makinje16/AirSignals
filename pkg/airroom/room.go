package airroom

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"strconv"

	"github.com/gorilla/websocket"
)

// AirRoom struct is the representation of a chat room
// in airWaves and keeps track of the number of clients
// in a room
type AirRoom struct {
	//the clients connected to this room
	AirClients *list.List
	//the unique id of the chatRoom
	ID string
	// waiting messages is used for when there is only a single user
	// in the room and for example an offer has been made but no user to send it to yet
	waitingMessages    *list.List
	numClients         int
	isNextClientPolite bool
}

// NewRoom creates an instance of a new Room with the specified
// ID and the first client
func NewRoom(c *AirClient, id string) *AirRoom {
	var room *AirRoom = new(AirRoom)
	room.ID = id
	room.numClients = 0
	room.AirClients = list.New()
	room.waitingMessages = list.New()
	room.ConnectClient(c)
	room.isNextClientPolite = true
	return room
}

// DisconnectUser removes the specified user from the room and decriments the user count
func (r *AirRoom) DisconnectUser(clientID string) error {
	log.Println(fmt.Sprintf("Removing user %s from chatroom %s", clientID, r.ID))
	for c := r.AirClients.Front(); c != nil; c = c.Next() {
		if c.Value.(*AirClient).ID == clientID {
			r.AirClients.Remove(c)
			r.numClients--
		}
	}
	return nil
}

// BroadcastMessage relies the message(message) that was sent from the Client with BroadcasterID
// to all other clients connected to the room (r)
func (r *AirRoom) BroadcastMessage(broadcasterID string, message []byte) {
	// If less than 2 users we will add that message to the waitingMessages list
	if r.numClients < 2 {
		newMessage := make(map[string][]byte)
		newMessage[broadcasterID] = message
		r.waitingMessages.PushBack(newMessage)
		return
	}
	// numclient >= 2
	for e := r.AirClients.Front(); e != nil; e = e.Next() {
		if e.Value.(*AirClient).ID != broadcasterID {
			e.Value.(*AirClient).Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// GetNumClients returns number of clients connected to room r
func (r *AirRoom) GetNumClients() int {
	return r.numClients
}

// ConnectClient takes a client as a parameter and adds that client to the room
// from which this method was invoked
func (r *AirRoom) ConnectClient(c *AirClient) error {
	if r.numClients >= 2 {
		log.Println(fmt.Sprintf("Max number of clients reached in chatroom %s. %s tried to connect", r.ID, c.ID))
		return errors.New("Max number of clients already reached")
	}
	r.numClients++
	r.AirClients.PushFront(c)
	c.Conn.WriteMessage(websocket.TextMessage, []byte("{\"type\":\"polite\", \"body\":\"Hi "+c.ID+"! You connected to the server at chatID: "+r.ID+", \"polite\":"+strconv.FormatBool(r.isNextClientPolite)+"\"}"))
	if r.isNextClientPolite || r.numClients == 0 {
		r.flipPolite()
	}

	log.Println(fmt.Sprintf("Successfully added user %s to chatroom %s", c.ID, r.ID))
	// if it is 2 at the end of the function it means a user was just added
	// there may be waitingMessages
	if r.numClients == 2 {
		// not actually n^2 because each element only has 1 entry so still O(n)
		for e := r.waitingMessages.Front(); e != nil; e = e.Next() {
			for k, v := range e.Value.(map[string][]byte) {
				r.BroadcastMessage(k, v)
			}
		}
	}
	return nil
}

func (r *AirRoom) flipPolite() {
	r.isNextClientPolite = !r.isNextClientPolite
}
