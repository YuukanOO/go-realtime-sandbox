package main

import (
	_ "embed"
	"errors"
	"gorealtimesandbox/realtime"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
)

//go:embed index.html
var indexSource string
var index = template.Must(template.New("").Parse(indexSource))

func main() {
	realtimeServer := realtime.NewServer()
	realtimeServer.Start()

	defer realtimeServer.Stop()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := index.Execute(w, nil); err != nil {
			log.Print(err)
		}
	})

	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		data, err := io.ReadAll(r.Body)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		realtimeServer.Send(string(data))
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		client, err := realtime.NewWebsocketClient(w, r, realtimeServer)

		if err != nil {
			log.Print(err)
			return
		}

		if err = client.Run(); err != nil {
			log.Print(err)
		}
	})

	mux.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		client, err := realtime.NewServerSentEventsClient(w, r, realtimeServer)

		if err != nil {
			log.Print(err)
			return
		}

		if err = client.Run(); err != nil {
			log.Print(err)
		}
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	exit := make(chan os.Signal, 1)

	signal.Notify(exit, os.Interrupt)

	go func() {
		log.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-exit
}
