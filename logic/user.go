package logic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

//系统用户
var System = &User{}
var globalUID uint32 = 1

type User struct {
	UID            int       `json:"uid"`
	Nickname       string    `json:"nickname"`
	EnterAt        time.Time `json:"enter_at"`
	Addr           string    `json:"addr"`
	messageChannel chan *Message
	Token          string `json:"token"`
	IsNew          bool   `json:"is_new"`
	conn           *websocket.Conn
}

func NewUser(conn *websocket.Conn, token, nickname string, addr string) *User {
	user := &User{
		Nickname:       nickname,
		EnterAt:        time.Now(),
		Addr:           addr,
		messageChannel: make(chan *Message, 8),
		Token:          token,
		conn:           conn,
	}
	if user.Token != "" && user.Token != "undefined" {

		uid, err := parseTokenAndValidate(token, nickname)
		if err == nil {
			user.UID = uid
		}
	}

	if user.UID == 0 {
		user.UID = int(atomic.AddUint32(&globalUID, 1))
		user.Token = genToken(user.UID, user.Nickname)
		user.IsNew = true
	}
	return user
}

func genToken(uid int, nickname string) string {
	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)
	messageMAC := macSha256([]byte(message), []byte(secret))

	return fmt.Sprintf("%suid%d", base64.StdEncoding.EncodeToString(messageMAC), uid)
}

func macSha256(message, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(message)
	return mac.Sum(nil)
}

func parseTokenAndValidate(token, nickname string) (int, error) {
	pos := strings.LastIndex(token, "uid")
	messageMAC, err := base64.StdEncoding.DecodeString(token[:pos])
	if err != nil {
		return 0, err
	}
	uid := cast.ToInt(token[pos+3:])

	secret := viper.GetString("token-secret")
	message := fmt.Sprintf("%s%s%d", nickname, secret, uid)
	ok := validateMAC([]byte(message), messageMAC, []byte(secret))
	if ok {
		return uid, nil
	}
	return 0, errors.New("token is illegal")
}

func validateMAC(message, messageMAC, secret []byte) bool {
	expectedMAC := macSha256(message, secret)
	return hmac.Equal(messageMAC, expectedMAC)
}

func (u *User) InMessageChannel(msg *Message) {
	u.messageChannel <- msg
}

func (u *User) SendMessage(ctx context.Context) {
	for msg := range u.messageChannel {
		_ = wsjson.Write(ctx, u.conn, msg)
	}
}

func (u *User) CloseMessageChannel() {
	close(u.messageChannel)
}

func (u *User) ReceiveMessage(ctx context.Context) error {
	var (
		receiveMsg map[string]string
		err        error
	)
	for {
		err = wsjson.Read(ctx, u.conn, &receiveMsg)
		if err != nil {
			//判断链接是否已关闭，如果正常关闭，则不认为是错误
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				return nil
			}
			return err
		}

		sendMsg := NewMessage(u, receiveMsg["content"])
		sendMsg.Content = FilterSenssitive(sendMsg.Content)
		reg := regexp.MustCompile(`@[^\s@]{2,20}`)
		sendMsg.Ats = reg.FindAllString(sendMsg.Content, -1)
		Broadcaster.Broadcast(sendMsg)
	}
}
