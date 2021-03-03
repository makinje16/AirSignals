package airroom

import (
	"container/list"
	"errors"
	"fmt"
	"log"
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
	waitingMessages *list.List
	numClients      int
	acceptingOffers bool
}

// NewRoom creates an instance of a new Room with the specified
// ID and the first client
func NewRoom(c *AirClient, id string) *AirRoom {
	var room *AirRoom = new(AirRoom)
	room.ID = id
	room.numClients = 0
	room.AirClients = list.New()
	room.waitingMessages = list.New()
	room.acceptingOffers = true
	room.ConnectClient(c)
	return room
}

// ConnectClient takes a client as a parameter and adds that client to the room
// from which this method was invoked
func (r *AirRoom) ConnectClient(c *AirClient) error {
	if r.numClients >= 2 {
		log.Println(fmt.Sprintf("Max number of clients reached in chatroom %s. %s tried to connect", r.ID, c.ID))
		return errors.New("Max number of clients already reached")
	}
	r.AirClients.PushFront(c)
	r.numClients++

	log.Println(fmt.Sprintf("Successfully added user %s to chatroom %s", c.ID, r.ID))
	log.Println(fmt.Sprintf("There are now %d users in the room", r.numClients))

	// if it is 2 at the end of the function it means a user was just added
	// there may be waitingMessages
	if r.numClients > 1 {
		r.PushQueue(c)
	}
	return nil
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
	if r.numClients < 2 {
		log.Println("Flushing Message Queue")
		r.FlushMessageQueue()
		r.acceptOffers()
	}
	return nil
}

// BroadcastMessage relays the message(message) that was sent from the Client with BroadcasterID
// to all other clients connected to the room (r)
func (r *AirRoom) BroadcastMessage(message *AirMessage) error {
	if message.MessageType == ClientOFFER && r.IsAcceptingOffers() {
		log.Println("Got an offer and changing room to not accept anymore")
		r.dontAcceptOffers()
	} else if message.MessageType == ClientOFFER && !r.IsAcceptingOffers() {
		return fmt.Errorf("AirRoom %s is currently not taking offers", r.ID)
	}

	// If less than 2 users we will add that message to the waitingMessages list
	if r.numClients < 2 {
		log.Println("Adding Message to back of the queue")
		r.waitingMessages.PushBack(message)
		return nil
	}

	// numclient >= 2
	for e := r.AirClients.Front(); e != nil; e = e.Next() {
		if e.Value.(*AirClient).ID != message.SenderID {
			log.Println(fmt.Sprintf("Sending Message of type %s from %s to %s", message.MessageType, message.SenderID, e.Value.(*AirClient).ID))
			e.Value.(*AirClient).SendMessage(message)
		}
	}
	return nil
}

// PushQueue takes all messages in the waitingMessages and sends them to the AirClient
// over its websocket connection
func (r *AirRoom) PushQueue(client *AirClient) {
	log.Println(fmt.Sprintf("Pushing Queue to %s", client.ID))
	for e := r.waitingMessages.Front(); e != nil; e = e.Next() {
		message := e.Value.(*AirMessage)
		log.Println(fmt.Sprintf("Sending Message of type %s from %s to %s", message.MessageType, message.SenderID, client.ID))
		client.SendMessage(message)
	}
}

// FlushMessageQueue empties the current list of waiting messages within the AirRoom struct
func (r *AirRoom) FlushMessageQueue() {
	r.waitingMessages.Init()
}

// AddToMessageQueue adds an AirMessage to the room (r) waitingMessages list
func (r *AirRoom) AddToMessageQueue(message *AirMessage) {
	r.waitingMessages.PushBack(message)
}

// GetNumClients returns number of clients connected to room r
func (r *AirRoom) GetNumClients() int {
	return r.numClients
}

// dontAcceptOffers changes the AirRoom to not accept incoming WebRTC Offers
func (r *AirRoom) dontAcceptOffers() {
	r.acceptingOffers = false
}

// acceptOffers changes the AirRoom to accept incoming WebRTC Offers
func (r *AirRoom) acceptOffers() {
	r.acceptingOffers = true
}

// IsAcceptingOffers returns whether the AirRoom instance is accepting WebRTC Offers
func (r *AirRoom) IsAcceptingOffers() bool {
	return r.acceptingOffers
}
