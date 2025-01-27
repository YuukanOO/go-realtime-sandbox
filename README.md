# go-realtime-sandbox

Tiny experiments with Golang, Server Sent Events and WebSockets.

## Build & Run

```sh
$ go run main.go
```

Go to [http://localhost:8080](http://localhost:8080) with multiple clients and choose the notification protocol.

Call the [http://localhost:8080/notify](http://localhost:8080/notify) endpoint (accepting every HTTP verb) to publish a message.
