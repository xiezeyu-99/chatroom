package main

import (
	"context"
	"fmt"

	"nhooyr.io/websocket"
)

func main() {
	c, _, err := websocket.Dial(context.Background(), "ws://localhost:2021/ws", nil)
	if err != nil {
		panic(err)
	}
	defer c.Close(websocket.StatusInternalError, "")

	fmt.Printf("链接成功")
}
