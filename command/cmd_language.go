package command

import (
	"fmt"
	"hashbot/database"
	"hashbot/types"
	"slices"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var SupportedLanguages = []string{"en", "pt"}

var language = Command{
	Name:              "language",
	Aliases:           []string{"lang"},
	Usage:             fmt.Sprintf("language [%s]", strings.Join(SupportedLanguages, "|")),
	Description:       "Set the bot's language for the sender",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) != 1 {
			return false
		}

		if !slices.Contains(SupportedLanguages, parsedArgs.Positional[0]) {
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

		err = database.UpdateUserLanguage(tx, message.Chatter.ID, parsedArgs.Positional[0])
		if err != nil {
			sender.Say(message.Channel, "❌Failed to set user's language", struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, "❌Failed to set user's language", struct {
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
				"Username": message.Chatter.Name,
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
