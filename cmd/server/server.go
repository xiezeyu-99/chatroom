package main

import (
	"log"
	"net/http"

	"nhooyr.io/websocket"
)

func main() {
	http.HandleFunc("/ws", func(w http.ResponseWriter, req *http.Request) {
		_, err := websocket.Accept(w, req, nil)
		if err != nil {
			panic(err)
		}

	})

	log.Fatal(http.ListenAndServe(":2021", nil))
}
