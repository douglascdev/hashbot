package command

import (
	"database/sql"
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var optin = Command{
	Name:        "optin",
	Aliases:     []string{},
	Usage:       "optin [all] | optin [command]",
	Description: "Opt in to one or all commands",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdOptionDescription",
				Other: "Opt in to one or all commands",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) != 1 {
			return false
		}
		return true
	},
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		var (
			fn func(tx *sql.Tx, userID string, optOut bool) error
			ok bool
		)
		if fn, ok = optoutOptions[args[1]]; !ok {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID: "UnknownCommand",
				},
				TemplateData: map[string]string{
					"Command": args[1],
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		err = fn(tx, message.Chatter.ID, false)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdOptedIn",
				Other: "✅ Opted in",
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		return nil
	},
}
