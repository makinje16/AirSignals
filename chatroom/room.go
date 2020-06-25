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
	room.numClients = 1
	room.Clients = list.New()
	room.Clients.PushFront(c)
	return room
}

func (r Room) broadcastMessage(c *Client, message string) (bool, error) {
	if r.numClients < 2 {
		return false, errors.New("no other client to broadcast message to")
	}
	for e := r.Clients.Front(); e != nil; e = e.Next() {
		if e.Value.(Client).ID != c.ID {
			e.Value.(*Client).Conn.WriteMessage(websocket.TextMessage, []byte(message))
		}
	}
	return true, nil
}

func (r Room) connectClient(c *Client) error {
	if r.numClients >= 2 {
		return errors.New("Max number of clients already reached")
	}
	r.numClients++
	r.Clients.PushFront(c)
	return nil
}
