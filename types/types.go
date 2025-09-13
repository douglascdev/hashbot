package types

import (
	"database/sql"
	"time"

	"monkebot/config"

	"github.com/gempir/go-twitch-irc/v4"
)

type Chatter struct {
	Name string
	ID   string

	IsMod         bool
	IsVIP         bool
	IsBroadcaster bool
}

// Message normalized to be platform agnostic
type Message struct {
	ID      string
	Message string
	Time    time.Time
	Channel string
	Cfg     *config.Config
	RoomID  string
	Chatter Chatter
	DB      *sql.DB
}

func NewMessage(msg twitch.PrivateMessage, db *sql.DB, cfg *config.Config) *Message {
	return &Message{
		ID:      msg.ID,
		Message: msg.Message,
		Time:    msg.Time,
		Channel: msg.Channel,
		RoomID:  msg.RoomID,
		Cfg:     cfg,
		Chatter: Chatter{
			Name:          msg.User.Name,
			ID:            msg.User.ID,
			IsMod:         msg.User.IsMod,
			IsVIP:         msg.User.IsVip,
			IsBroadcaster: msg.User.IsBroadcaster,
		},
		DB: db,
	}
}

type SenderParam int

const (
	ReplyMessageID SenderParam = iota
	Me
)

type MessageSender interface {
	Say(channel string, message string, params ...struct {
		Param SenderParam
		Value string
	})

	Join(channels ...string)
	Part(channels ...string)
	Ping() (time.Duration, error)

	Uptime() time.Duration

	Buttify(message string) string
	ShouldButtify() bool
}
