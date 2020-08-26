package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/makinje16/AirSignals/pkg/airroom"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

var localhostflag bool

var threadSafeRooms = struct {
	sync.RWMutex
	chatRooms map[string]*airroom.AirRoom
}{chatRooms: make(map[string]*airroom.AirRoom)}

func checkOrigin(r *http.Request) bool {
	return true
}

func main() {
	flag.BoolVar(&localhostflag, "localhost", true, "true if running on localhost false if on public ip")
	flag.Parse()

	router := gin.Default()
	router.GET("/ws/:chatID/:hostID", socket)
	if localhostflag {
		router.Run("localhost:8080")
	}
	router.Run("0.0.0.0:8080")
}

func socket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	chatID := c.Param("chatID")
	hostID := c.Param("hostID")
	conn.WriteMessage(websocket.TextMessage, []byte("Hi "+hostID+"! You connected to the server at chatID: "+chatID))

	// Lock the chatRooms map to modify data
	threadSafeRooms.RWMutex.Lock()
	_, ok := threadSafeRooms.chatRooms[chatID]

	// Means someone is already in the chatroom
	if ok {
		err := threadSafeRooms.chatRooms[chatID].ConnectClient(airroom.NewClient(hostID, conn))
		if err != nil {
			log.Println(err)
		}
	} else {
		// First person to be in the chatroom
		threadSafeRooms.chatRooms[chatID] = airroom.NewRoom(airroom.NewClient(hostID, conn), chatID)
	}
	threadSafeRooms.RWMutex.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		if messageType == websocket.TextMessage {
			fmt.Println(p)
			airMessage := airroom.NewEmptyAirMessage()
			json.Unmarshal(p, airMessage)

			threadSafeRooms.RWMutex.Lock()
			switch messageType := airMessage.MessageType; messageType {
			case airroom.ClientOFFER:
				log.Println("Offer")
			case airroom.ClientANSWER:
				log.Println("Answer")
			case airroom.ClientCANDIDATE:
				log.Println("Candidate")
			case airroom.ClientMESSAGE:
				log.Println("Message")
			}

			// conn.WriteMessage(websocket.TextMessage, p)
			// threadSafeRooms.chatRooms[chatID].BroadcastMessage(hostID, p)
			threadSafeRooms.RWMutex.Unlock()
		}

	}

}
