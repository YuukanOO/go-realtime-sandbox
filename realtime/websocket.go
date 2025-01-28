package realtime

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type wsClient struct {
	conn   *websocket.Conn
	send   chan string
	server Server
}

func NewWebsocketClient(w http.ResponseWriter, r *http.Request, server Server) (Client, error) {
	c, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	return &wsClient{
		conn:   c,
		send:   make(chan string),
		server: server,
	}, nil
}

func (w *wsClient) Run() (err error) {
	if err = w.server.Register(w); err != nil {
		return err
	}

	go w.readPump()
	go w.writePump()

	return nil
}

func (w *wsClient) Send(message string) {
	w.send <- message
}

func (w *wsClient) readPump() {
	defer w.close()

	for {
		_, message, err := w.conn.ReadMessage() // Will return an error when the connection is closed.

		if err != nil {
			return
		}

		w.server.Send(string(message)) // Broadcast the message to all clients
	}
}

func (w *wsClient) writePump() {
	defer w.conn.Close() // Only close the connection here, let the readPump unregister the client

	for msg := range w.send {
		err := w.conn.WriteMessage(websocket.TextMessage, []byte(msg))

		if err != nil {
			return
		}
	}
}

func (w *wsClient) close() {
	close(w.send)
	_ = w.conn.Close()
	w.server.UnRegister(w)
}
