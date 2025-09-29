package hashbot

import (
	"database/sql"
	"errors"
	"fmt"
	"hashbot/command"
	"hashbot/config"
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
	"slices"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/douglascdev/buttifier"
	"github.com/rs/zerolog/log"
	"golang.org/x/text/language"

	"github.com/Potat-Industries/go-potatFilters"
	"github.com/gempir/go-twitch-irc/v4"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type HashBot struct {
	TwitchClient       *twitch.Client
	Cfg                *config.Config
	db                 *sql.DB
	bundle             *i18n.Bundle
	invalidatedTokenCh chan bool
	startTime          time.Time
	buttifier          *buttifier.Buttifier
}

func NewHashBot(cfg *config.Config, db *sql.DB, invalidatedTokenCh chan bool) (*HashBot, error) {
	client := twitch.NewClient(cfg.Login, "oauth:"+cfg.TwitchToken)

	butt, err := buttifier.New(buttifier.English)
	butt.ButtificationProbability = 0.05
	butt.ButtificationRate = 0.2
	if err != nil {
		return nil, fmt.Errorf("failed to initialize buttifier: %w", err)
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("active.pt.toml")

	mb := &HashBot{
		TwitchClient:       client,
		Cfg:                cfg,
		db:                 db,
		bundle:             bundle,
		invalidatedTokenCh: invalidatedTokenCh,
		startTime:          time.Now(),
		buttifier:          butt,
	}

	mb.BindClientFunctions()

	return mb, nil
}

func (t *HashBot) BindClientFunctions() {
	client := t.TwitchClient

	if client == nil {
		return
	}

	db := t.db
	cfg := t.Cfg
	bundle := t.bundle
	mb := t
	invalidatedTokenCh := t.invalidatedTokenCh

	client.OnPrivateMessage(func(message twitch.PrivateMessage) {
		startTime := time.Now()
		normalizedMsg := types.NewMessage(message, db, cfg)
		tx, err := db.Begin()
		if err != nil {
			log.Err(err).Msg("failed to start transaction")
			return
		}
		defer tx.Rollback()
		userLanguage, err := database.SelectUserLanguage(tx, normalizedMsg.Chatter.ID)
		var localizer *i18n.Localizer
		if err != nil {
			localizer = i18n.NewLocalizer(bundle, "en")
			normalizedMsg.UserLang = "en"
			log.Warn().Err(err).Str("channel", message.Channel).Str("user", message.User.Name).Msg("failed to select user's language")
		} else {
			localizer = i18n.NewLocalizer(bundle, userLanguage)
			normalizedMsg.UserLang = userLanguage
		}
		normalizedMsg.Localizer = localizer

		err = command.HandleCommands(normalizedMsg, mb, cfg)
		if errors.Is(err, command.UnknownCommandErr) {
			log.Warn().Str("user", message.User.Name).Str("msg", message.Message).Msg("unknown command")
			mb.Say(message.Channel, "❌Unknown command", struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return
		}
		if err != nil {
			log.Err(err).Msg("command failed")
			// try to refresh our token if twitch returns a 401
			if strings.Contains(err.Error(), "401") {
				log.Warn().Msg("401 is in command fail error message, invalidating token")
				invalidatedTokenCh <- true
			}
		}
		internalLatency := fmt.Sprintf("%d ms", time.Since(startTime).Milliseconds())
		log.Info().
			Str("channel", message.Channel).
			Str("user", message.User.Name).
			Str("msg", message.Message).
			Str("internalLatency", internalLatency).
			Msg("new message")
	})

	client.OnConnect(func() {
		log.Info().
			Str("login", cfg.Login).
			Msg("connected to Twitch")

		tx, err := db.Begin()
		if err != nil {
			log.Err(err).Msg("failed to initialize transaction")
		}
		defer tx.Rollback()

		// initial inserts are done, just join saved channels
		if idToNameMap, err := database.SelectJoinedChannels(tx); err == nil && len(idToNameMap) > 0 {
			var userIds []string
			for id := range idToNameMap {
				userIds = append(userIds, id)
			}
			// get updated names from the api, in case an user changed usernames
			updatedIdToUser, err := twitchapi.GetUserByID(cfg, userIds...)
			if err != nil {
				log.Err(err).Msg("failed to get user ids")
				return
			}

			changed := make(map[string]string)
			for updatedId, updatedUser := range updatedIdToUser {
				if oldName, found := idToNameMap[updatedId]; found && oldName != updatedUser.Login {
					changed[updatedId] = updatedUser.Login
				}
				idToNameMap[updatedId] = updatedUser.Login
			}

			if len(changed) > 0 {
				log.Info().Int("count", len(changed)).Msg("username changes detected")
			}

			err = database.UpdateUsernames(tx, changed)
			if err != nil {
				log.Err(err).Msg("failed to update usernames")
				return
			}

			var updatedUsernames []string
			for _, name := range idToNameMap {
				updatedUsernames = append(updatedUsernames, name)
			}
			log.Info().Strs("usernames", updatedUsernames).Msg("updated usernames")

			mb.Join(updatedUsernames...)
			log.Info().Strs("channels", updatedUsernames).Msg("successfully joined saved channels")
			return
		} else if err != nil {
			log.Err(err)
		}

		mb.Join(cfg.InitialChannels...)

		var cmdNames []string
		for _, cmd := range command.Commands {
			cmdNames = append(cmdNames, cmd.Name)
		}

		err = database.InsertCommands(tx, cmdNames...)
		if err != nil {
			log.Err(err).Msg("failed to insert commands")
			return
		}

		var helixUsers []twitchapi.HelixUser
		helixUsers, err = twitchapi.GetUserByName(cfg, cfg.InitialChannels...)
		if err != nil {
			log.Err(err).Strs("channels", cfg.InitialChannels).Msg("failed to get helix data for users")
			return
		}

		var users []struct {
			ID   string
			Name string
		}
		for _, twitchUser := range helixUsers {
			users = append(users, struct {
				ID   string
				Name string
			}{twitchUser.ID, twitchUser.Login})
		}

		err = database.InsertUsers(tx, true, "en", users...)
		if err != nil {
			log.Err(err).Msg("failed to insert initial users in the database")
			return
		}

		for _, user := range users {
			if !slices.Contains(cfg.AdminUsernames, user.Name) {
				continue
			}
			err = database.UpdateUserPermission(tx, user.Name, "admin")
			if err != nil {
				log.Err(err).Str("name", user.Name).Str("id", user.ID).Msg("failed to insert user commands for user")
				return
			}
		}

		for _, twitchUser := range helixUsers {
			err = database.InsertUserCommands(tx, twitchUser.ID, cmdNames...)
			if err != nil {
				log.Err(err).Str("name", twitchUser.Login).Str("id", twitchUser.ID).Msg("failed to insert user commands for user")
				return
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Err(err).Msg("failed to commit transaction")
			return
		}

		log.Info().Msg("successfully inserted initial channels")
	})

	client.OnSelfJoinMessage(func(message twitch.UserJoinMessage) {
		log.Info().Str("channel", message.Channel).Msg("joined channel")
	})

	client.OnSelfPartMessage(func(message twitch.UserPartMessage) {
		log.Info().Str("channel", message.Channel).Msg("parted channel")
	})
}

func (t *HashBot) Connect() error {
	return t.TwitchClient.Connect()
}

func (t *HashBot) Join(channels ...string) {
	t.TwitchClient.Join(channels...)
}

func (t *HashBot) Part(channels ...string) {
	for _, channel := range channels {
		t.TwitchClient.Depart(channel)
	}
}

func (t *HashBot) Say(channel string, message string, params ...struct {
	Param types.SenderParam
	Value string
},
) {
	if message == "" {
		log.Warn().Msg("ignored attempt to send empty message")
		return
	}

	// read params
	var (
		replyMessageID string
		me             bool
	)
	for _, param := range params {
		switch param.Param {
		case types.Me:
			me = param.Value == "true"
		case types.ReplyMessageID:
			replyMessageID = param.Value
		}
	}

	// filter banned phrases
	if potatFilters.Test(message, potatFilters.FilterStrict) {
		log.Warn().
			Str("channel", channel).
			Str("msg", message).
			Msg("message filtered")
		t.TwitchClient.Say(channel, "⚠ Message withheld for containing a banned phrase...")
		return
	}

	// send response
	var response strings.Builder
	if me {
		const meStr = "/me "
		response.WriteString(meStr)
	}

	const invisPrefix = "󠀀 " // prevents command injection
	response.WriteString(invisPrefix)

	response.WriteString(message)

	s := response.String()

	if replyMessageID != "" {
		log.Debug().Str("channel", channel).Str("replyMessageID", replyMessageID).Str("msg", s).Msg("replying")
		t.TwitchClient.Reply(channel, replyMessageID, s)
		return
	}

	log.Debug().Str("channel", channel).Str("msg", s).Msg("sending message")
	t.TwitchClient.Say(channel, s)
}

func (t *HashBot) Ping() (duration time.Duration, err error) {
	duration, err = t.TwitchClient.Latency()
	return
}

func (t *HashBot) Uptime() time.Duration {
	return time.Since(t.startTime)
}

func (t *HashBot) ShouldButtify() bool {
	return t.buttifier.ToButtOrNotToButt()
}

func (t *HashBot) Buttify(message string) string {
	return t.buttifier.ButtifySentence(message)
}
