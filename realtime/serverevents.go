package realtime

import (
	"fmt"
	"net/http"
)

type sseClient struct {
	writer  http.ResponseWriter
	request *http.Request
	send    chan string
	server  Server
}

func NewServerSentEventsClient(w http.ResponseWriter, r *http.Request, server Server) (Client, error) {
	return &sseClient{
		writer:  w,
		request: r,
		server:  server,
		send:    make(chan string),
	}, nil
}

func (s *sseClient) Run() error {
	s.writer.Header().Set("Access-Control-Allow-Origin", "*")
	s.writer.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	s.writer.Header().Set("Content-Type", "text/event-stream")
	s.writer.Header().Set("Cache-Control", "no-cache")
	s.writer.Header().Set("Connection", "keep-alive")

	if err := s.server.Register(s); err != nil {
		return err
	}

	defer s.close()

	for {
		select {
		case <-s.request.Context().Done():
			return nil
		case msg, ok := <-s.send:
			if !ok {
				return nil
			}

			if _, err := fmt.Fprintf(s.writer, "data: %s\n\n", msg); err != nil {
				return err
			}

			s.writer.(http.Flusher).Flush()
		}
	}
}

// Send implements Client.
func (s *sseClient) Send(msg string) {
	s.send <- msg
}

func (s *sseClient) close() {
	close(s.send)
	s.server.UnRegister(s)
}
