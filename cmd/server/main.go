package main

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type (
	// Message describes the data structure of a message object
	Message struct {
		Text      string    `json:"text"`
		Timestamp time.Time `json:"timestamp"`
		SenderID  string    `json:"senderid"`
		Code      string    `json:"code"`
	}

	room struct {
		names            map[string]int
		clients          map[string]*websocket.Conn
		addClientChan    chan *websocket.Conn
		removeClientChan chan *websocket.Conn
		broadcastChan    chan Message
		Messages         *ring.Ring
		cnt              int
	}
)

var (
	once        sync.Once
	port        = "9001"
	ran         = rand.New(rand.NewSource(time.Now().UnixNano()))
	storageSize = 5
)

const (
	storageFilename = "./cmd/data.json"
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
		Messages:         ring.New(storageSize),
		// Counts the actual number of elements
		// in Messages as ring may not be full
		cnt: 0,
	}
}

// Reads data from json file into destination v
func readStorage(v interface{}, fname string) {
	jsonFile, err := os.Open(fname)
	if err != nil {
		log.Printf("Error opening file %v : %v", fname, err.Error())
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	if err = json.Unmarshal(byteValue, v); err != nil {
		log.Println("Error reading from storage:", err.Error())
	}
}

// Initializes room storage data
func (rm *room) initStorage() {
	_, err := os.Stat(storageFilename)
	if os.IsNotExist(err) {
		// Create new file
		os.Create(storageFilename)

	} else {
		// Copy data from storage file to runtime
		var storedM []Message
		readStorage(&storedM, storageFilename)
		for _, m := range storedM {
			rm.Messages.Value = m
			rm.Messages = rm.Messages.Next()
			rm.cnt++
		}
	}
}

// Initializes room paramters
func initRoom() *room {
	rm := newRoom()
	rm.initStorage()
	// go rm.run()
	return rm
}

// Adds a new websocket connection to the server
func (rm *room) addClient(conn *websocket.Conn) {
	rm.clients[conn.RemoteAddr().String()] = conn
	log.Println("New client connected:", conn.RemoteAddr().String())

	// Acknowledge client's entry
	name := rm.findUniqueName()
	websocket.JSON.Send(conn, Message{Text: name, Code: entryCode})

	// Send last n messages to client, starting from oldest
	r := rm.Messages
	if rm.cnt < storageSize {
		for c := 0; c < rm.cnt; c++ {
			r = r.Prev()
		}
	}
	for c := 0; c < rm.cnt; c++ {
		websocket.JSON.Send(conn, r.Value)
		r = r.Next()
	}
}

// Removes websocket connection from server
func (rm *room) removeClient(conn *websocket.Conn) {
	log.Println("Client disconnected:", conn.RemoteAddr().String())
	websocket.JSON.Send(conn, Message{Code: exitCode})
	delete(rm.clients, conn.RemoteAddr().String())
}

// Broadcast message to all connected clients
func (rm *room) broadcast(m Message) {
	for _, conn := range rm.clients {
		if err := websocket.JSON.Send(conn, m); err != nil {
			log.Println("Error broadcasting")
		}
	}
}

// Adds message m to ring data structure
func (rm *room) store(m Message) {
	rm.Messages.Value = m
	rm.Messages = rm.Messages.Next()
	if rm.cnt < rm.Messages.Len() {
		rm.cnt++
	}
}

// Writes to storage so that messages even after termination
func (rm *room) writeToStorage(filename string) {
	// Copies ring messages into an array
	ms := make([]Message, 0)
	r := rm.Messages
	r.Do(func(m interface{}) {
		if m != nil {
			ms = append(ms, m.(Message))
		}
	})

	// Writes array into json file
	file, err := json.MarshalIndent(ms, "", " ")
	if err != nil {
		log.Println("Json marshal error")
	}

	if err = ioutil.WriteFile(filename, file, 0644); err != nil {
		log.Println("Json write error", err.Error())
	}
}

// Runs the main functions of the chatroom
func (rm *room) run() {
	// singleton
	for {
		select {
		case conn := <-rm.addClientChan:
			rm.addClient(conn)
		case conn := <-rm.removeClientChan:
			rm.removeClient(conn)
			rm.writeToStorage(storageFilename)
		case m := <-rm.broadcastChan:
			rm.store(m)
			rm.broadcast(m)
		}
	}
}

// Handler for each client connection
func handler(ws *websocket.Conn, rm *room) {
	go once.Do(rm.run)
	rm.addClientChan <- ws

	var m Message
	for {
		err := websocket.JSON.Receive(ws, &m)

		// client exit sequence
		if m.Text == "/exit" || err != nil {
			rm.removeClientChan <- ws
			return
		}
		rm.broadcastChan <- m
	}
}

// Connects a new room the given port number
func connect(port string) error {
	rm := initRoom()

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
	if err := connect(port); err != nil {
		log.Println(err.Error())
	}
}
