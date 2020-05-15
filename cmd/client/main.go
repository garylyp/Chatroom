package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

type (
	// Data structure of message
	Message struct {
		Text      string    `json:"text"`
		Timestamp time.Time `json:"timestamp"`
		SenderID  string    `json:"senderid"`
		Code      string    `json:"code"`
	}
)

var (
	port = "9000"
	id   string
)

const (
	entryCode   = "ENTRY_OK"
	exitCode    = "EXIT_OK"
	defaultCode = "OK"
	divider     = "===================="
)

// Generates a random unique IP for client to simulate multiple clients on local machine
func generatedIP() string {
	var arr [4]int
	for i := 0; i < 4; i++ {
		rand.Seed(time.Now().UnixNano())
		arr[i] = rand.Intn(256)
	}
	id = fmt.Sprintf("http://%d.%d.%d.%d", arr[0], arr[1], arr[2], arr[3])
	return id
}

// Connects to the server via websocket
func connect(port string) (*websocket.Conn, error) {
	return websocket.Dial(fmt.Sprintf("ws://localhost:%s", port), "", generatedIP())
}

// Receives message from server
func receive(ws *websocket.Conn) {
	var m Message
	for {
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println("You have disconnected")
			return
		}

		switch m.Code {
		case entryCode:
			// Receive assigned name from server upon successful entry
			id = m.Text
			fmt.Printf("Welcome, %v. To quit, enter \"/exit\"\n", id)
			continue
		case exitCode:
			// Receive exit ack from server
			return
		default:
			// Receive message from other chatroom users
			fmt.Print(prettifyMsg(m))
		}

	}
}

// Adds border around message content to denote a textbox
func prettifyMsg(m Message) string {
	return fmt.Sprintf("%v\n%v [%v]: \n%v\n%v\n",
		divider,
		m.SenderID,
		m.Timestamp.Format("02-01 15:04"),
		m.Text,
		divider)
}

// Sends a message to the chatroom
func send(ws *websocket.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()

		switch text {
		case "":
			continue
		case "/exit":
			// Sends special command to exit chatroom
			return
		default:
			m := Message{
				Text:      text,
				Timestamp: time.Now(),
				SenderID:  id,
				Code:      defaultCode,
			}
			err := websocket.JSON.Send(ws, m)
			if err != nil {
				log.Println("Error sending message: ", err.Error())
				break
			}
		}
	}
}

func main() {

	// connect
	ws, err := connect(port)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// receive
	go receive(ws)

	// send
	send(ws)
}
