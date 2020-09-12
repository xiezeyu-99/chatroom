package logic

import (
	"container/ring"

	"github.com/spf13/viper"
)

//离线消息
type offlineProcessor struct {
	n int

	//保存所有用户的最近n条消息
	recentRing *ring.Ring

	//保存某个用户的最近n条消息
	userRing map[string]*ring.Ring
}

var OfflineProcessor = newOfflineProcessor()

func newOfflineProcessor() *offlineProcessor {
	n := viper.GetInt("offline-num")

	return &offlineProcessor{
		n:          n,
		recentRing: ring.New(n),
		userRing:   make(map[string]*ring.Ring),
	}
}

//存离线消息
func (o *offlineProcessor) Save(msg *Message) {
	if msg.Type != MsgTypeNormal {
		return
	}

	o.recentRing.Value = msg
	o.recentRing = o.recentRing.Next()

	for _, nickname := range msg.Ats {
		nickname = nickname[1:]
		var (
			r  *ring.Ring
			ok bool
		)
		if r, ok = o.userRing[nickname]; !ok {
			r = ring.New(o.n)
		}
		r.Value = msg
		o.userRing[nickname] = r.Next()
	}
}

func (o *offlineProcessor) Send(user *User) {
	o.recentRing.Do(func(value interface{}) {
		if value != nil {
			user.InMessageChannel(value.(*Message))
		}
	})

	if user.IsNew {
		return
	}

	if r, ok := o.userRing[user.Nickname]; ok {
		r.Do(func(value interface{}) {
			if value != nil {
				user.InMessageChannel(value.(*Message))
			}
		})
		delete(o.userRing, user.Nickname)
	}
}
