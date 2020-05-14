package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

type (
	user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	Message struct {
		Text      string    `json:"text"`
		Timestamp time.Time `json:"timestamp"`
		SenderID  string    `json:"senderid"`
	}
)

var (
	port = flag.String("port", "9000", "port used for ws connection")
	id   string
)

func connect() (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://localhost:%s", *port), "", generatedIP())
}

// Generates a random IP to differentiate users while testing locally
func generatedIP() string {
	var arr [4]int
	for i := 0; i < 4; i++ {
		rand.Seed(time.Now().UnixNano())
		arr[i] = rand.Intn(256)
	}
	id = fmt.Sprintf("http://%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
	return id
}

func receive(ws *websocket.Conn) {
	var m Message
	for {
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println("Error receiving message: ", err.Error())
			break
		}
		fmt.Printf("%v [%v]: \n%v\n", m.SenderID, m.Timestamp.Format("02-01 15:04"), m.Text)
	}

	// Display message

}

func send(ws *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}

		m := Message{
			Text:      text,
			Timestamp: time.Now(),
			SenderID:  id,
		}
		err := websocket.JSON.Send(ws, m)
		if err != nil {
			log.Println("Error sending message: ", err.Error())
			break
		}
	}
}

func main() {
	flag.Parse()

	// connect
	ws, err := connect()
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// receive
	go receive(ws)

	// send
	send(ws)
}
