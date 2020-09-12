package main

import (
	"chatroom/server"
	"fmt"
	"log"
	"net/http"

	"chatroom/global"
)

var (
	addr   = ":2022"
	banner = `Go编程之旅————一起用GO左项目：ChatRoom,start on：%s`
)

func init() {
	global.Init()
}

func main() {
	fmt.Printf(banner+"\n", addr)
	server.RegisterHandle()

	log.Fatal(http.ListenAndServe(addr, nil))
}
