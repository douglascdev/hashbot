package command

import (
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var setLevel = Command{
	Name:        "setlevel",
	Aliases:     []string{"permission", "perm", "level"},
	Usage:       "setlevel [username] [permission]",
	Description: "Set a user's permission level",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdSetLevelDescription",
				Other: "Set a user's permission level",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) != 2 {
			return false
		}
		return true
	},
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		var isAdmin bool
		isAdmin, err = database.SelectIsUserAdmin(tx, message.Chatter.ID)
		if err != nil {
			return err
		}
		if !isAdmin {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "MustBeAdmin",
					Other: "❌You must be an admin to use this command",
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		var userExists bool
		userExists, err = database.SelectUserExists(tx, args[1])
		if err != nil {
			return err
		}

		if !userExists {
			var users []twitchapi.HelixUser
			users, err = twitchapi.GetUserByName(message.Cfg, args[1])
			if err != nil {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID: "UserNotFound",
					},
					TemplateData: map[string]string{
						"User": args[1],
					},
				})
				sender.Say(message.Channel, msg, struct {
					Param types.SenderParam
					Value string
				}{Param: types.ReplyMessageID, Value: message.ID})
				return err
			}
			user := users[0]
			// user isn't in the db but exists on twitch, so it's a new user
			err = database.InsertUsers(tx, false, "en", struct{ ID, Name string }{user.ID, user.Login})
			if err != nil {
				return err
			}
		}

		err = database.UpdateUserPermission(tx, args[1], args[2])
		if err != nil {
			return err
		}
		err = tx.Commit()
		if err != nil {
			return err
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "UpdatedPermissions",
				Other: "✅ Updated {{.User}}'s permission to {{.Permission}}!",
			},
			TemplateData: map[string]string{
				"User":       args[1],
				"Permission": args[2],
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		log.Info().Str("channel", message.Channel).Str("user", message.Chatter.Name).Str("permission", args[2]).Msg("successfully updated user permission")

		return nil
	},
}
