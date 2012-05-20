package pubsuber

import (
	"../redis"
	"../trigger"
	"io"
)

func userMessageChannel(user string) string {
	return "Users:" + user
}

func urlMessageChannel(url string) string {
	return "Url:" + url
}

func userInstanceIndex(user string) string {
	return "User Instances " + user
}

func messageReceiver(channel string, action func(string)) (io.Closer, error) {
	t := trigger.New()
	closeChannel := t.Channel()

	sub, err := redis.Subscribe(channel)
	if err != nil {
		return nil, err
	}

	go func() {
		defer sub.Close()
		defer sub.Unsubscribe(channel)

		for {
			select {
			case m := <-sub.Messages:
				action(m.Elem.String())
			case <-closeChannel:
				return
			}
		}
	}()

	return t, nil
}

type Target interface {
	SendMessage(message string)
	ReceiveMessages(action func(message string)) io.Closer
	IsActive() bool
}

type UserTarget interface {
}

func User(name string) Target {
	return userTarget{name}
}

type userTarget struct {
	user string
}

func (this userTarget) SendMessage(message string) {
	if this.IsActive() {
		redis.Client.Publish(userMessageChannel(this.user), message)
	} else {
		redis.Client.Lpush(userMessageChannel(this.user), message)
	}
}

func (this userTarget) ReceiveMessages(action func(message string)) io.Closer {
	closer, _ := messageReceiver(userMessageChannel(this.user), action)
	redis.Client.Incr(userInstanceIndex(this.user))

	oldmessagecount, _ := redis.Client.Llen(userMessageChannel(this.user))
	var i int64
	for i = 0; i < oldmessagecount; i++ {
		message, _ := redis.Client.Rpop(userMessageChannel(this.user))
		redis.Client.Publish(userMessageChannel(this.user), message.String())
	}

	return trigger.OnClose(func() {
		closer.Close()
		redis.Client.Decr(userInstanceIndex(this.user))
	})
}

func (this userTarget) IsActive() bool {
	count, _ := redis.Client.Get(userInstanceIndex(this.user))
	return count.Int64() > 0
}

func Url(url string) Target {
	return urlTarget{url}
}

type urlTarget struct {
	url string
}

func (this urlTarget) SendMessage(message string) {
	redis.Client.Publish(urlMessageChannel(this.url), message)
}

func (this urlTarget) ReceiveMessages(action func(message string)) io.Closer {
	closer, _ := messageReceiver(urlMessageChannel(this.url), action)
	return closer
}

func (this urlTarget) IsActive() bool {
	return true
}