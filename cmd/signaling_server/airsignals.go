package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
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

// TO-DO: find a way to have the RWMutex on individual AirRooms
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

	// Setup Logging
	file, err := os.OpenFile("/var/log/airsignals.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Println("AirSignals starting...")

	// Setup Gin router and middleware
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
	router.GET("/ws/:chatID/:hostID", socket)
	router.GET("/getConnectedClients/:chatID", checkClients)

	// Run signaling server
	if localhostflag {
		router.Run("localhost:8080")
	}
	router.Run("0.0.0.0:8080")
}

func checkClients(c *gin.Context) {
	chatID := c.Param("chatID")
	_, ok := threadSafeRooms.chatRooms[chatID]
	if !ok {
		c.JSON(401, gin.H{
			"type": "message",
			"body": "Chat Room does not exist",
		})
	} else {
		c.JSON(200, gin.H{
			"type":       "message",
			"numClients": threadSafeRooms.chatRooms[chatID].GetNumClients(),
			"body":       "",
		})
	}
}

func socket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	chatID := c.Param("chatID")
	hostID := c.Param("hostID")
	conn.SetCloseHandler(func(code int, text string) error {

		threadSafeRooms.RWMutex.Lock()
		_, ok := threadSafeRooms.chatRooms[chatID]
		if ok {
			threadSafeRooms.chatRooms[chatID].DisconnectUser(hostID)
		}
		threadSafeRooms.RWMutex.Unlock()
		log.Println("Removed user " + hostID + " from room " + chatID)
		return nil
	})

	log.Println(fmt.Sprintf("%s is attempting to connected to the server at chatid: %s", hostID, chatID))

	// Lock the chatRooms map to modify data
	threadSafeRooms.RWMutex.Lock()
	_, ok := threadSafeRooms.chatRooms[chatID]

	// Means someone is already in the chatroom
	if ok {
		err := threadSafeRooms.chatRooms[chatID].ConnectClient(airroom.NewClient(hostID, conn))
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Connected " + hostID + " to room " + chatID)
			log.Println("Room " + chatID + " now has " + strconv.Itoa(threadSafeRooms.chatRooms[chatID].GetNumClients()) + " user(s) connected")
		}
	} else {
		// First person to be in the chatroom
		threadSafeRooms.chatRooms[chatID] = airroom.NewRoom(airroom.NewClient(hostID, conn), chatID)
		log.Println("Created room " + chatID + " and connected " + hostID)
	}
	threadSafeRooms.RWMutex.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		if messageType == websocket.TextMessage {
			airMessage := airroom.NewEmptyAirMessage()
			json.Unmarshal(p, airMessage)
			threadSafeRooms.RWMutex.Lock()
			err := threadSafeRooms.chatRooms[chatID].BroadcastMessage(airMessage)
			if err != nil {
				log.Println(fmt.Sprintf("Message not broadcasted: %s\n\tmessage: %v", err.Error(), airMessage))
			}
			threadSafeRooms.RWMutex.Unlock()
		}

	}

}
