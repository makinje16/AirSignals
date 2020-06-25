package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

var localhostflag bool

func checkOrigin(r *http.Request) bool {
	return true
}

func main() {
	flag.BoolVar(&localhostflag, "localhost", true, "true if running on localhost false if on public ip")
	flag.Parse()

	router := gin.Default()
	router.GET("/ws", socket)

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

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
		}

		if messageType == websocket.TextMessage {
			fmt.Println(string(p))
			conn.WriteMessage(websocket.TextMessage, []byte("Hello Client!"))
		}
	}

}
