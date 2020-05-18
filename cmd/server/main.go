package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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

	// Data structure of stored messages
	Storage struct {
		Head     int             `json:"head"`
		Idx      int             `json:"idx"`
		Messages map[int]Message `json:"messages"`
	}

	room struct {
		names            map[string]int
		clients          map[string]*websocket.Conn
		addClientChan    chan *websocket.Conn
		removeClientChan chan *websocket.Conn
		broadcastChan    chan Message
		data             Storage
	}
)

var (
	port                   = "9000"
	ran         *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	storageSize            = 5
)

const (
	storageFilename = "./internal/storage/data.json"
	entryCode       = "ENTRY_OK"
	exitCode        = "EXIT_OK"
	defaultCode     = "OK"
)

// Generate unique name 5 character name that is not in use
func (rm *room) findUniqueName() string {
	var name string
	ok := true
	for ok {
		name = fmt.Sprintf("u%04d", ran.Intn(9999))
		_, ok = rm.clients[name]
	}
	return name
}

// Initializes a new chatroom
func newRoom() *room {
	return &room{
		names:            make(map[string]int),
		clients:          make(map[string]*websocket.Conn),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan Message),
	}
}

// Initializes room storage data
func (rm *room) initStorage() {
	_, err := os.Stat(storageFilename)

	if os.IsNotExist(err) {
		os.Create(storageFilename)
		rm.data = Storage{
			Head:     0,
			Idx:      0,
			Messages: make(map[int]Message),
		}
	} else {
		jsonFile, err := os.Open(storageFilename)
		defer jsonFile.Close()

		if err != nil {
			log.Printf("Error opening file %v : %v", storageFilename, err.Error())
		}

		byteValue, _ := ioutil.ReadAll(jsonFile)

		err = json.Unmarshal(byteValue, &rm.data)
		if err != nil {
			log.Println("Error reading from storage:", err.Error())
		}
	}
}

// Add a new websocket connection to the server
func (rm *room) addClient(conn *websocket.Conn) {
	rm.clients[conn.RemoteAddr().String()] = conn
	log.Println("New client connected:", conn.RemoteAddr().String())

	// Acknowledge client's entry
	name := rm.findUniqueName()
	websocket.JSON.Send(conn, Message{Text: name, Code: entryCode})

	// Send last n messages to client, starting from oldest
	i := rm.data.Head
	n := len(rm.data.Messages)
	for c := 0; c < n; c++ {
		websocket.JSON.Send(conn, rm.data.Messages[i])
		i = (i + 1) % n
	}
}

// Removes websocket connection with client
func (rm *room) removeClient(conn *websocket.Conn) {
	log.Println("Client disconnected:", conn.RemoteAddr().String())
	websocket.JSON.Send(conn, Message{Code: exitCode})
	delete(rm.clients, conn.RemoteAddr().String())
}

// Broadcast message m to all connected websocket clients
func (rm *room) broadcast(m Message) {
	for _, conn := range rm.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			log.Println("Error broadcasting")
		}
	}
}

func (s *Storage) writeToStorage(filename string) {
	// Writes messages into storage file
	file, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		log.Println("Json marshal error")
	}
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		log.Println("Json write error", err.Error())
	}
}

// Adds message m to storage, purging oldest message if necessary
func (rm *room) store(m Message) {

	d := &rm.data
	// Adjust pointers in circular queue
	if len(d.Messages) == storageSize {
		d.Head = (d.Head + 1) % storageSize
	}
	d.Messages[d.Idx] = m
	d.Idx = (d.Idx + 1) % storageSize

	go d.writeToStorage(storageFilename)
}

// Runs the main functions of the chatroom
func (rm *room) run() {
	for {
		select {
		case conn := <-rm.addClientChan:
			rm.addClient(conn)
		case conn := <-rm.removeClientChan:
			rm.removeClient(conn)
		case m := <-rm.broadcastChan:
			rm.store(m)
			rm.broadcast(m)
		}
	}
}

// Handles a new incoming websocket connection
func handler(ws *websocket.Conn, rm *room) {
	go rm.run()

	rm.addClientChan <- ws

	for {
		var m Message
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			rm.removeClient(ws)
			return
		}
		rm.broadcastChan <- m
	}
}

// Connects a new room the given port number
func connect(port string) error {
	rm := newRoom()
	rm.initStorage()

	// init http request multiplexor
	mux := http.NewServeMux()

	// routes "/" to handler
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		handler(ws, rm)
	}))

	// set server parameters
	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	log.Println("Server running on port: ", port)
	return s.ListenAndServe()
}

func main() {

	err := connect(port)
	if err != nil {
		log.Println(err.Error())
	}

}
