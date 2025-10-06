package types

import (
	"database/sql"
	"time"

	"hashbot/config"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
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
	ID        string
	Message   string
	Time      time.Time
	Channel   string
	Cfg       *config.Config
	RoomID    string
	Chatter   Chatter
	DB        *sql.DB
	Localizer *i18n.Localizer
	UserLang  string
}

var SupportedLanguages = []string{"pt", "en"}

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

type UserLocation struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Latitude    float64  `json:"latitude"`
	Longitude   float64  `json:"longitude"`
	Elevation   float64  `json:"elevation"`
	FeatureCode string   `json:"feature_code"`
	CountryCode string   `json:"country_code"`
	Admin1ID    int      `json:"admin1_id"`
	Timezone    string   `json:"timezone"`
	Population  int      `json:"population"`
	Postcodes   []string `json:"postcodes"`
	CountryID   int      `json:"country_id"`
	Country     string   `json:"country"`
	Admin1      string   `json:"admin1"`
}

type FindLocationResult struct {
	Results          []UserLocation `json:"results"`
	GenerationtimeMs float64        `json:"generationtime_ms"`
}
