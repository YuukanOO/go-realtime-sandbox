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

	go func() {
		defer w.close()

		for {
			_, _, err := w.conn.ReadMessage()

			if err != nil {
				break
			}
		}
	}()

	go func() {
		defer w.conn.Close()

		for msg := range w.send {
			err = w.conn.WriteMessage(websocket.TextMessage, []byte(msg))

			if err != nil {
				return
			}
		}
	}()

	return nil
}

func (w *wsClient) Send(message string) {
	w.send <- message
}

func (w *wsClient) close() {
	close(w.send)
	w.server.UnRegister(w)
	_ = w.conn.Close()
}
