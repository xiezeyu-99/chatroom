package server

import (
	"chatroom/logic"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func WebSocketHandleFunc(w http.ResponseWriter, req *http.Request) {
	conn, err := websocket.Accept(w, req, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Println("websocket accept error:", err)
		return
	}

	nickname := req.FormValue("nickname")
	token := req.FormValue("token")
	if l := len(nickname); l < 4 || l > 20 {
		log.Println("nickname illegal:", nickname)
		_ = wsjson.Write(req.Context(), conn, logic.NewErrorMessage("非法昵称，昵称长度：4-20"))
		conn.Close(websocket.StatusUnsupportedData, "nickname illegal!")
		return
	}

	if !logic.Broadcaster.CanEnterRoom(nickname) {
		log.Println("昵称已存在：", nickname)
		_ = wsjson.Write(req.Context(), conn, logic.NewErrorMessage("昵称已存在"))
		conn.Close(websocket.StatusUnsupportedData, "nickname exists!")
		return
	}

	//新用户进来，构建实例
	user := logic.NewUser(conn, token, nickname, req.RemoteAddr)

	//开启给用户发送消息的goroutine
	go user.SendMessage(req.Context())

	//给新用户发送欢迎信息
	user.InMessageChannel(logic.NewWelcomeMessage(user))

	//向所有用户告知新用户的到来
	msg := logic.NewUserEnterMessage(user)
	logic.Broadcaster.Broadcast(msg)

	//讲该用户加入广播器的用户列表
	logic.Broadcaster.UserEntering(user)
	log.Println("user:", nickname, "join chat")

	//接收用户消息
	err = user.ReceiveMessage(req.Context())

	//用户离开
	logic.Broadcaster.UserLeaving(user)
	msg = logic.NewUserLeaveMessage(user)
	logic.Broadcaster.Broadcast(msg)
	log.Println("user：", nickname, "leave chat")

	if err == nil {
		conn.Close(websocket.StatusNormalClosure, "")
	} else {
		log.Println("read from client error:", err)
		conn.Close(websocket.StatusInternalError, "read rom client error")
	}
}
