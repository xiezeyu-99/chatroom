package main

import (
	"chatroom/logic"
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "net/http/pprof"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	userNum       int
	loginInterval time.Duration //登录间隔
	msgInterval   time.Duration //发送消息间隔
)

func init() {
	flag.IntVar(&userNum, "u", 500, "登录用户数")
	flag.DurationVar(&loginInterval, "l", 5e9, "用户登录时间间隔")
	flag.DurationVar(&msgInterval, "m", 1*time.Minute, "用户发送消息间隔")
}

func main() {
	flag.Parse()

	for i := 0; i < userNum; i++ {
		go UserConnect("user" + strconv.Itoa(i))
		time.Sleep(loginInterval)
	}
	select {}
}

func UserConnect(nickname string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws://127.0.0.1:2022/ws?nickname="+nickname, nil)
	if err != nil {
		log.Println("Dial error:", err)
		return
	}
	defer conn.Close(websocket.StatusInternalError, "内部错误")

	go SendMessage(conn, nickname)

	ctx = context.Background()

	for {
		var message logic.Message
		err = wsjson.Read(ctx, conn, &message)
		if err != nil {
			log.Println("receive msg error:", err)
			continue
		}

		if message.ClientSendTime.IsZero() {
			continue
		}

		if d := time.Since(message.ClientSendTime); d > 1*time.Second {
			fmt.Printf("接收的服务端响应（%d）: %#v\n", d.Milliseconds(), message)
		}
	}
	// conn.Close(websocket.StatusNormalClosure, "")
}

func SendMessage(conn *websocket.Conn, nickname string) {
	ctx := context.Background()
	i := 1
	for {
		msg := map[string]string{
			"content":   "来自" + nickname + "的消息：" + strconv.Itoa(i),
			"send_time": strconv.FormatInt(time.Now().UnixNano(), 10),
		}
		err := wsjson.Write(ctx, conn, msg)
		if err != nil {
			log.Println("send msg error:", err, "nickname:", nickname, "no:", i)
		}
		i++

		time.Sleep(msgInterval)
	}
}
