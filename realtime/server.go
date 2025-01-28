package realtime

import (
	"errors"
	"log"
	"sync"
)

type serverState uint8

const (
	stopped serverState = iota
	stopping
	started
)

var ErrInvalidOperation = errors.New("invalid operation given the state of the server")

type Server interface {
	Start() error
	Stop() error
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
	// Close a client, called by the server when stopping
	Close()
}

type server struct {
	state        serverState
	muState      sync.Mutex
	clients      []Client
	wg           sync.WaitGroup // WaitGroup used to wait for all clients to be disconnected when shutting down
	exiting      chan struct{}
	connected    chan Client
	disconnected chan Client
	messages     chan string
}

func NewServer() Server {
	return &server{}
}

func (s *server) Start() error {
	if err := s.tryMovingToState(started); err != nil {
		return err
	}

	s.exiting = make(chan struct{})
	s.connected = make(chan Client)
	s.disconnected = make(chan Client)
	s.messages = make(chan string)

	log.Print("starting realtime server")

	go s.run()
	return nil
}

func (s *server) Stop() error {
	if err := s.tryMovingToState(stopping); err != nil {
		return err
	}

	log.Print("exiting realtime server, waiting for all clients to be disconnected...")
	s.exiting <- struct{}{}

	s.wg.Wait()

	// Close all channels
	close(s.exiting)
	close(s.connected)
	close(s.disconnected)
	close(s.messages)

	s.clients = nil

	log.Print("exited realtime server")
	return s.tryMovingToState(stopped)
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

func (s *server) run() {
	exiting := false // Manage this bool here to avoid managing concurrency

	for {
		select {
		case _, ok := <-s.exiting:
			// Channel closed, let's exit the for loop!
			if !ok {
				return
			}

			exiting = true

			for _, c := range s.clients {
				c.Close()
			}
		case msg := <-s.messages:
			for _, c := range s.clients {
				c.Send(msg)
			}
		case client := <-s.connected:
			// If the server is stopping, do not accept new clients
			if exiting {
				client.Close()
				break
			}

			s.clients = append(s.clients, client)
			s.wg.Add(1)
			log.Printf("client connected %T (%d left)", client, len(s.clients))
		case client := <-s.disconnected:
			for i, c := range s.clients {
				if c != client {
					continue
				}

				s.clients[i] = nil
				s.clients = append(s.clients[:i], s.clients[i+1:]...)
				s.wg.Done()
				log.Printf("client disconnected %T (%d left)", client, len(s.clients))
			}
		}
	}
}

func (s *server) tryMovingToState(newState serverState) error {
	s.muState.Lock()
	defer s.muState.Unlock()

	if s.state == newState {
		return nil
	}

	if (newState == started && s.state != stopped) ||
		(newState == stopping && s.state != started) ||
		(newState == stopped && s.state != stopping) {
		return ErrInvalidOperation
	}

	s.state = newState
	return nil
}
