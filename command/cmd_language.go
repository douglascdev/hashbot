package command

import (
	"fmt"
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
	"slices"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var language = Command{
	Name:              "language",
	Aliases:           []string{"lang"},
	Usage:             fmt.Sprintf("language [%s]", strings.Join(types.SupportedLanguages, "|")),
	Description:       "Set the bot's language for the author or a specified channel",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) != 1 {
			return false
		}

		if !slices.Contains(types.SupportedLanguages, parsedArgs.Positional[0]) {
			return false
		}

		return true
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		failMsg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SetUserLanguageFail",
				Other: "❌Failed to set user's language",
			},
		})

		targetID, targetUsername := message.Chatter.ID, message.Chatter.Name
		if len(parsedArgs.HashPrefixed) > 0 {
			targetUsername = parsedArgs.HashPrefixed[0]
			users, err := twitchapi.GetUserByName(message.Cfg, targetUsername)
			if err != nil {
				sender.Say(message.Channel, failMsg, struct {
					Param types.SenderParam
					Value string
				}{types.ReplyMessageID, message.ID})
				return err
			}

			if len(users) == 0 {
				userNotFound := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "UserNotFound",
						Other: "❌User '{{.User}}' not found",
					},
					TemplateData: map[string]string{
						"User": targetUsername,
					},
				})
				sender.Say(message.Channel, userNotFound, struct {
					Param types.SenderParam
					Value string
				}{types.ReplyMessageID, message.ID})
				return err
			}
			targetID = users[0].ID

			isAdmin, err := database.SelectIsUserAdmin(tx, message.Chatter.ID)
			if err != nil {
				sender.Say(message.Channel, failMsg, struct {
					Param types.SenderParam
					Value string
				}{types.ReplyMessageID, message.ID})
				return err
			}

			if targetID != message.Chatter.ID && !isAdmin {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "CmdButtwordDisallowed",
						Other: "❌You must be an admin or the channel's owner to change this setting",
					},
				})
				sender.Say(message.Channel, msg, struct {
					Param types.SenderParam
					Value string
				}{types.ReplyMessageID, message.ID})
				return err
			}
		}

		err = database.UpdateUserLanguage(tx, targetID, parsedArgs.Positional[0])
		if err != nil {
			sender.Say(message.Channel, failMsg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, failMsg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SetUserLanguageResult",
				Other: "✅Set {{.Username}}'s language to {{.Language}}",
			},
			TemplateData: map[string]string{
				"Username": targetUsername,
				"Language": parsedArgs.Positional[0],
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})

		return nil
	},
}
