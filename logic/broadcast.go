package logic

import (
	"log"

	"github.com/spf13/viper"
)

//广播器
type broadcaster struct {
	users map[string]*User

	enteringChannel chan *User
	leavingChannel  chan *User
	messageChannel  chan *Message

	checkUserChannel      chan string
	checkUserCanInChannel chan bool
	// 获取用户列表
	requestUsersChannel chan struct{}
	usersChannel        chan []*User
}

var Broadcaster = &broadcaster{
	users:           make(map[string]*User),
	enteringChannel: make(chan *User),
	leavingChannel:  make(chan *User),
	messageChannel:  make(chan *Message, viper.GetInt("message-channel-buffer")),

	checkUserChannel:      make(chan string),
	checkUserCanInChannel: make(chan bool),

	// 获取用户列表
	requestUsersChannel: make(chan struct{}),
	usersChannel:        make(chan []*User),
}

//启动广播
func (b *broadcaster) Start() {
	for {
		select {
		case user := <-b.enteringChannel:
			b.users[user.Nickname] = user
			OfflineProcessor.Send(user)
		case user := <-b.leavingChannel:
			delete(b.users, user.Nickname)
			user.CloseMessageChannel()
		case msg := <-b.messageChannel:
			//广播消息
			for _, user := range b.users {
				if user.UID == msg.User.UID {
					continue
				}
				user.InMessageChannel(msg)
			}
			OfflineProcessor.Save(msg)
		case nickname := <-b.checkUserChannel:
			if _, ok := b.users[nickname]; ok {
				b.checkUserCanInChannel <- false
			} else {
				b.checkUserCanInChannel <- true
			}
		case <-b.requestUsersChannel:
			userList := make([]*User, 0, len(b.users))
			for _, user := range b.users {
				userList = append(userList, user)
			}

			b.usersChannel <- userList
		}
	}
}

//广播消息
func (b *broadcaster) Broadcast(msg *Message) {
	if len(b.messageChannel) >= viper.GetInt("message-channel-buffer") {
		log.Println("broadcast queue 满了")
	}
	b.messageChannel <- msg
}

func (b *broadcaster) CanEnterRoom(nickname string) bool {
	b.checkUserChannel <- nickname
	return <-b.checkUserCanInChannel
}

func (b *broadcaster) UserEntering(user *User) {
	b.enteringChannel <- user
}

func (b *broadcaster) UserLeaving(user *User) {
	b.leavingChannel <- user
}

func (b *broadcaster) GetUserList() []*User {
	b.requestUsersChannel <- struct{}{}
	return <-b.usersChannel
}
