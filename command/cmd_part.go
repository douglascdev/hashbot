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

var part = Command{
	Name:        "part",
	Aliases:     []string{"leave"},
	Usage:       "part | part [channel]",
	Description: "Leave the message author's channel or the specified channel",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdPartDescription",
				Other: "Leave the message author's channel or the specified channel",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var channelsToLeave []struct {
			ID   string
			Name string
		}

		if len(args) == 2 && message.Chatter.Name == args[1] {
			channelsToLeave = append(channelsToLeave, struct {
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
			channelsToLeave = make([]struct {
				ID   string
				Name string
			}, 0, len(twitchUsers))

			for _, user := range twitchUsers {
				channelsToLeave = append(channelsToLeave, struct {
					ID   string
					Name string
				}{ID: user.ID, Name: user.Login})
			}
		} else {
			channelsToLeave = append(channelsToLeave, struct {
				ID   string
				Name string
			}{ID: message.Chatter.ID, Name: message.Chatter.Name})
		}

		if len(channelsToLeave) == 0 {
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
		query = fmt.Sprintf("SELECT name FROM user WHERE name IN (%s) AND bot_is_joined", strings.Repeat("?,", max(0, len(channelsToLeave)-1))+"?")
		channels = make([]interface{}, len(channelsToLeave))
		for i, channel := range channelsToLeave {
			channels[i] = channel.Name
		}

		rows, err = tx.Query(query, channels...)
		if err != nil {
			return err
		}

		defer rows.Close()
		foundChannels := map[string]struct{}{}
		for rows.Next() {
			var name string
			err = rows.Scan(&name)
			if err != nil {
				return err
			}
			foundChannels[name] = struct{}{}
		}
		err = rows.Err()
		if err != nil {
			return err
		}

		if len(foundChannels) != len(channelsToLeave) {
			channelsNotFound := make([]string, 0, len(channelsToLeave)-len(foundChannels))
			for _, channel := range channelsToLeave {
				if _, ok := foundChannels[channel.Name]; !ok {
					channelsNotFound = append(channelsNotFound, channel.Name)
				}
			}
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdPartNotJoined",
					Other: "❌The following channels were not joined: {{.Channels}}",
				},
				TemplateData: map[string]any{
					"Channels": strings.Join(channelsNotFound, ", "),
				},
			})
			sender.Say(message.Channel, msg)
			return nil
		}

		// ensure all joined channels have bot_is_joined set to false if InsertUsers didn't just insert them(it skips existing users)
		var channelIDs []string
		for _, channel := range channelsToLeave {
			channelIDs = append(channelIDs, channel.ID)
		}
		err = database.UpdateIsBotJoined(tx, false, channelIDs...)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		channelNames := make([]string, len(channelsToLeave))
		for i, channel := range channelsToLeave {
			channelNames[i] = channel.Name
		}
		log.Info().Strs("channels", channelNames).Msg("successfully parted channels")
		sender.Part(channelNames...)

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdPartResult",
				Other: "✅Successfully parted {{.Channels}}",
			},
			TemplateData: map[string]any{
				"Channels": strings.Join(channelNames, ", "),
			},
		})
		sender.Say(message.Channel, msg)
		return nil
	},
}
