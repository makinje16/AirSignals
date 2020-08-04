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
	ID string
	// waiting messages is used for when there is only a single user
	// in the room and for example an offer has been made but no user to send it to yet
	waitingMessages *list.List
	numClients      int
}

// NewRoom creates an instance of a new Room with the specified
// ID and the first client
func NewRoom(c *Client, id string) *Room {
	var room *Room = new(Room)
	room.ID = id
	room.numClients = 0
	room.Clients = list.New()
	room.waitingMessages = list.New()
	room.ConnectClient(c)
	// fmt.Println("Created New Room and connected client " + c.ID)
	// fmt.Println("Room " + room.ID + " now has " + strconv.Itoa(room.numClients) + " user(s) connected")
	return room
}

// DisconnectUser removes the specified user from the room and decriments the user count
func (r *Room) DisconnectUser(clientID string) error {
	for c := r.Clients.Front(); c != nil; c = c.Next() {
		if c.Value.(*Client).ID == clientID {
			r.Clients.Remove(c)
			r.numClients--
		}
	}
	return nil
}

// BroadcastMessage relies the message(message) that was sent from the Client with BroadcasterID
// to all other clients connected to the room (r)
func (r *Room) BroadcastMessage(broadcasterID string, message []byte) {
	// If less than 2 users we will add that message to the waitingMessages list
	if r.numClients < 2 {
		newMessage := make(map[string][]byte)
		newMessage[broadcasterID] = message
		r.waitingMessages.PushBack(newMessage)
		return
	}
	// numclient >= 2
	for e := r.Clients.Front(); e != nil; e = e.Next() {
		if e.Value.(*Client).ID != broadcasterID {
			e.Value.(*Client).Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// GetNumClients returns number of clients connected to room r
func (r *Room) GetNumClients() int {
	return r.numClients
}

// ConnectClient takes a client as a parameter and adds that client to the room
// from which this method was invoked
func (r *Room) ConnectClient(c *Client) error {
	if r.numClients >= 2 {
		return errors.New("Max number of clients already reached")
	}
	r.numClients++
	r.Clients.PushFront(c)

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
	// fmt.Println("Added Client")
	// fmt.Println("Room " + r.ID + " now has " + strconv.Itoa(r.numClients) + " connected clients")
	// fmt.Println(c)
	return nil
}
