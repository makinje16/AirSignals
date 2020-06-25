package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/makinje16/AirSignals/chatroom"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

var localhostflag bool

var threadSafeRooms = struct {
	sync.RWMutex
	chatRooms map[string]*chatroom.Room
}{chatRooms: make(map[string]*chatroom.Room)}

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
	if ok {
		err := threadSafeRooms.chatRooms[chatID].ConnectClient(chatroom.NewClient(hostID, conn))
		if err != nil {
			log.Println(err)
		}
	} else {
		threadSafeRooms.chatRooms[chatID] = chatroom.NewRoom(chatroom.NewClient(hostID, conn), chatID)
	}
	threadSafeRooms.RWMutex.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		if messageType == websocket.TextMessage {
			fmt.Println(p)
			threadSafeRooms.RWMutex.Lock()
			threadSafeRooms.chatRooms[chatID].BroadcastMessage(hostID, p)
			threadSafeRooms.RWMutex.Unlock()
		}

	}

}
