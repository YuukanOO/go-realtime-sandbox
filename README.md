# go-realtime-sandbox

Tiny experiments with Golang, Server Sent Events and WebSockets.

## Build & Run

```sh
$ go run main.go
```

Go to [http://localhost:8080](http://localhost:8080) with multiple clients and choose the notification protocol.

Call the [http://localhost:8080/send](http://localhost:8080/send) endpoint (using the POST method) to publish a message.
