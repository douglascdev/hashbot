package command

import (
	"hashbot/database"
	"hashbot/types"
	"slices"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var allowedButtWords = []string{
	"butt",
	"glorp",
}

var buttword = Command{
	Name:              "buttword",
	Aliases:           []string{},
	Usage:             "buttword [butt|glorp]",
	Description:       "Set word used by buttsbot in the user's channel to replace syllables",
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

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtwordFailed",
				Other: "❌Failed to set buttword",
			},
		})
		if err != nil {
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		err = database.UpdateUserButtword(tx, message.Chatter.ID, parsedArgs.Positional[0])
		if err != nil {
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		err = tx.Commit()
		if err != nil {
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		msg = message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtwordSuccess",
				Other: "Successfully set user's buttword to {{.Word}}",
			},
			TemplateData: map[string]any{
				"Word": parsedArgs.Positional[0],
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})

		return nil
	},
}
