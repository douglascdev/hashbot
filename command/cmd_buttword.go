package command

import (
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
	"slices"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var allowedButtWords = []string{
	"butt",
	"glorp",
}

var buttword = Command{
	Name:        "buttword",
	Aliases:     []string{"bw"},
	Usage:       "buttword [butt|glorp] | buttword [butt|glorp] #channel",
	Description: "Set word used by buttsbot in the author's channel or the specified channel to replace syllables",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtwordDescription",
				Other: "Set word used by buttsbot in the author's channel or the specified channel to replace syllables",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) != 1 {
			return false
		}
		if !slices.Contains(allowedButtWords, parsedArgs.Positional[0]) {
			return false
		}
		return true
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()

		failMsg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtwordFailed",
				Other: "❌Failed to set buttword",
			},
		})
		if err != nil {
			sender.Say(message.Channel, failMsg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

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
						ID: "UserNotFound",
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
						ID:    "MustBeAdmin",
						Other: "❌You must be an admin to use this command",
					},
				})
				sender.Say(message.Channel, msg, struct {
					Param types.SenderParam
					Value string
				}{types.ReplyMessageID, message.ID})
				return err
			}
		}
		err = database.UpdateUserButtword(tx, targetID, parsedArgs.Positional[0])
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

		failMsg = message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtwordSuccess",
				Other: "Successfully set {{.User}}'s channel buttword to {{.Word}}",
			},
			TemplateData: map[string]any{
				"Word": parsedArgs.Positional[0],
				"User": targetUsername,
			},
		})
		sender.Say(message.Channel, failMsg, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})

		return nil
	},
}
