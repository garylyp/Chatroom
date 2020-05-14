package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type (
	Message struct {
		Text      string
		Timestamp time.Time `json:"timestamp"`
		SenderID  string    `json:"senderid"`
	}

	room struct {
		name             string
		clients          map[string]*websocket.Conn
		addClientChan    chan *websocket.Conn
		removeClientChan chan *websocket.Conn
		broadcastChan    chan Message
	}
)

var (
	port = flag.String("port", "9000", "port used for ws connection")
)

func newRoom() *room {
	return &room{
		name:             "Main Room",
		clients:          make(map[string]*websocket.Conn),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan Message),
	}
}

func (rm *room) addClient(conn *websocket.Conn) {
	rm.clients[conn.RemoteAddr().String()] = conn
}

func (rm *room) removeClient(conn *websocket.Conn) {
	delete(rm.clients, conn.RemoteAddr().String())
}

func (rm *room) broadcast(m Message) {
	for _, conn := range rm.clients {
		err := websocket.JSON.Send(conn, m)
		if err != nil {
			log.Println("Error in broadcasting")
		}
	}
}

func (rm *room) run() {
	for {
		select {
		case conn := <-rm.addClientChan:
			rm.addClient(conn)
		case conn := <-rm.removeClientChan:
			rm.removeClient(conn)
		case m := <-rm.broadcastChan:
			rm.broadcast(m)
		}
	}
}

func handler(ws *websocket.Conn, rm *room) {
	go rm.run()

	rm.addClientChan <- ws
	// polling
	for {
		var m Message
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println("Removing disconnected client")
			rm.removeClient(ws)
			return
		}
		rm.broadcastChan <- m
	}
}

func connect(port string) error {
	rm := newRoom()
	// init http request multiplexor
	mux := http.NewServeMux()
	// routes "/" to handler
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		handler(ws, rm)
	}))

	s := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	return s.ListenAndServe()
}

func main() {
	// Hello world, the web server
	flag.Parse()
	connect(*port)

}
