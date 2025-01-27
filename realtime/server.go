package realtime

import (
	"log"
)

type Server interface {
	Start()
	Stop()
	Send(string)

	// Methods used by clients

	Register(Client) error
	UnRegister(Client)
}

type Client interface {
	// Starts the client loop to process messages.
	Run() error
	// Send a raw message to the client.
	Send(string)
}

type server struct {
	clients      []Client
	exiting      chan struct{}
	connected    chan Client
	disconnected chan Client
	messages     chan string
}

func NewServer() Server {
	return &server{
		clients:      make([]Client, 0),
		connected:    make(chan Client),
		disconnected: make(chan Client),
		exiting:      make(chan struct{}),
		messages:     make(chan string),
	}
}

func (s *server) Start() {
	log.Print("starting realtime server")
	go s.loop()
}

func (s *server) Stop() {
	close(s.exiting)
	log.Print("exiting realtime server")
}

func (s *server) Send(message string) {
	s.messages <- message
}

func (s *server) Register(client Client) error {
	s.connected <- client
	return nil
}

func (s *server) UnRegister(client Client) {
	s.disconnected <- client
}

func (s *server) loop() {
	for {
		select {
		case <-s.exiting:
			return
		case msg := <-s.messages:
			for _, c := range s.clients {
				c.Send(msg)
			}
		case client := <-s.connected:
			s.clients = append(s.clients, client)
			log.Printf("client connected %T (%d left)", client, len(s.clients))
		case client := <-s.disconnected:
			for i, c := range s.clients {
				if c != client {
					continue
				}

				s.clients = append(s.clients[:i], s.clients[i+1:]...)
				log.Printf("client disconnected %T (%d left)", client, len(s.clients))
			}
		}
	}
}
