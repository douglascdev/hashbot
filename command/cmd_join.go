package command

import (
	"database/sql"
	"fmt"
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var join = Command{
	Name:              "join",
	Aliases:           []string{},
	Usage:             "join | join [channel]",
	Description:       "Join the message author's channel or the specified channel",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdJoinDescription",
				Other: "Join the message author's channel or the specified channel",
			},
		})
	},
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var channelsToJoin []struct {
			ID   string
			Name string
		}

		if len(args) == 2 && message.Chatter.Name == args[1] {
			channelsToJoin = append(channelsToJoin, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		} else if len(args) > 1 {
			isAdmin := false
			isAdmin, err = database.SelectIsUserAdmin(tx, message.Chatter.ID)

			if err != nil && err != sql.ErrNoRows {
				return err
			}

			if err == sql.ErrNoRows || !isAdmin {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "MustBeAdmin",
						Other: "❌You must be an admin to use this command",
					},
				})
				sender.Say(message.Channel, msg)
				return nil
			}

			var twitchUsers []twitchapi.HelixUser
			twitchUsers, err = twitchapi.GetUserByName(message.Cfg, args[1:]...)
			if err != nil {
				return err
			}
			channelsToJoin = make([]struct {
				ID   string
				Name string
			}, 0, len(twitchUsers))

			for _, user := range twitchUsers {
				channelsToJoin = append(channelsToJoin, struct {
					ID   string
					Name string
				}{ID: user.ID, Name: user.Login})
			}
		} else {
			channelsToJoin = append(channelsToJoin, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		}

		if len(channelsToJoin) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "ChannelsNotFound",
					Other: "❌Channel(s) not found",
				},
			})
			sender.Say(message.Channel, msg)
			return nil
		}

		// check if any of the channels are already in the database
		var (
			query    string
			rows     *sql.Rows
			channels []interface{}
		)
		query = fmt.Sprintf("SELECT name FROM user WHERE name IN (%s) AND bot_is_joined", strings.Repeat("?,", max(0, len(channelsToJoin)-1))+"?")
		channels = make([]interface{}, len(channelsToJoin))
		for i, channel := range channelsToJoin {
			channels[i] = channel.Name
		}

		rows, err = tx.Query(query, channels...)
		if err == nil {
			defer rows.Close()
			var foundChannels []string
			for rows.Next() {
				var name string
				err = rows.Scan(&name)
				if err != nil {
					return err
				}
				foundChannels = append(foundChannels, name)
			}
			err = rows.Err()
			if err != nil {
				return err
			}

			if len(foundChannels) > 0 {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "ChannelsAlreadyJoined",
						Other: "❌The following channels were already joined: {{.Channels}}",
					},
					TemplateData: map[string]any{
						"Channels": strings.Join(foundChannels, ", "),
					},
				})
				sender.Say(message.Channel, msg)
				return nil
			}
		}

		err = database.InsertUsers(tx, true, message.UserLang, channelsToJoin...)
		if err != nil {
			return err
		}

		// ensure all joined channels have bot_is_joined set to true if InsertUsers didn't just insert them(it skips existing users)
		var channelIDs []string
		for _, channel := range channelsToJoin {
			channelIDs = append(channelIDs, channel.ID)
		}
		err = database.UpdateIsBotJoined(tx, true, channelIDs...)
		if err != nil {
			return err
		}

		var commandNames []string
		rows, err = tx.Query("SELECT name FROM command")
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return err
			}
			commandNames = append(commandNames, name)
		}

		for _, channel := range channelsToJoin {
			err = database.InsertUserCommands(tx, channel.ID, commandNames...)
			if err != nil {
				log.Warn().Err(err).Str("channel", channel.Name).Msg("failed to insert user commands after join, skipping channel")
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		channelNames := make([]string, len(channelsToJoin))
		for i, channel := range channelsToJoin {
			channelNames[i] = channel.Name
		}
		log.Info().Strs("channels", channelNames).Msg("successfully joined channels")
		sender.Join(channelNames...)

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdJoinResult",
				One:   "✅Joined channel {{.Channels}}",
				Other: "✅Joined channels {{.Channels}}",
			},
			PluralCount: len(channelNames),
			TemplateData: map[string]any{
				"Channels": strings.Join(channelNames, ", "),
			},
		})
		sender.Say(message.Channel, msg)
		for _, channel := range channelsToJoin {
			sender.Say(channel.Name, "ola")
		}
		return nil
	},
}
