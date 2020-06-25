package chatroom

import (
	"container/list"
	"errors"

	"github.com/gorilla/websocket"
)

// Room struct is the representation of a chat room
// in airWaves and keeps track of the number of clients
// in a room
type Room struct {
	//the clients connected to this room
	Clients *list.List
	//the unique id of the chatRoom
	ID         string
	numClients int
}

// NewRoom creates an instance of a new Room with the specified
// ID and the first client
func NewRoom(c *Client, id string) *Room {
	var room *Room = new(Room)
	room.ID = id
	room.numClients = 0
	room.Clients = list.New()
	room.ConnectClient(c)
	// fmt.Println("Created New Room and connected client " + c.ID)
	// fmt.Println("Room " + room.ID + " now has " + strconv.Itoa(room.numClients) + " user(s) connected")
	return room
}

// BroadcastMessage relies the message(message) that was sent from the Client with BroadcasterID
// to all other clients connected to the room (r)
func (r *Room) BroadcastMessage(broadcasterID string, message []byte) (bool, error) {
	if r.numClients < 2 {
		return false, errors.New("no other client to broadcast message to")
	}
	for e := r.Clients.Front(); e != nil; e = e.Next() {
		if e.Value.(*Client).ID != broadcasterID {
			e.Value.(*Client).Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
	return true, nil
}

// ConnectClient takes a client as a parameter and adds that client to the room
// from which this method was invoked
func (r *Room) ConnectClient(c *Client) error {
	if r.numClients >= 2 {
		return errors.New("Max number of clients already reached")
	}
	r.numClients++
	r.Clients.PushFront(c)
	// fmt.Println("Added Client")
	// fmt.Println("Room " + r.ID + " now has " + strconv.Itoa(r.numClients) + " connected clients")
	// fmt.Println(c)
	return nil
}
