package types

import (
	"database/sql"
	"errors"
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

type GetWeatherResult struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	GenerationtimeMs     float64 `json:"generationtime_ms"`
	UtcOffsetSeconds     int     `json:"utc_offset_seconds"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Elevation            float64 `json:"elevation"`
	CurrentUnits         struct {
		Time                string `json:"time"`
		Interval            string `json:"interval"`
		ApparentTemperature string `json:"apparent_temperature"`
		Precipitation       string `json:"precipitation"`
		Rain                string `json:"rain"`
		Snowfall            string `json:"snowfall"`
		WindSpeed10M        string `json:"wind_speed_10m"`
		CloudCover          string `json:"cloud_cover"`
		Temperature2M       string `json:"temperature_2m"`
	} `json:"current_units"`
	Current struct {
		Time                string  `json:"time"`
		Interval            int     `json:"interval"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		Precipitation       float64 `json:"precipitation"`
		Rain                float64 `json:"rain"`
		Snowfall            float64 `json:"snowfall"`
		WindSpeed10M        float64 `json:"wind_speed_10m"`
		CloudCover          int     `json:"cloud_cover"`
		Temperature2M       float64 `json:"temperature_2m"`
	} `json:"current"`
}

type CommandError interface {
	error

	UserError() string
}

type commandError struct {
	err     error
	userErr string
}

func (c *commandError) UserError() string {
	return c.userErr
}

func (c *commandError) Error() string {
	return c.err.Error()
}

func NewCommandError(err error, userErr string) CommandError {
	if err == nil {
		err = errors.New("command error")
	}
	return &commandError{
		err:     err,
		userErr: userErr,
	}
}
