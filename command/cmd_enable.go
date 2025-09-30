package command

import (
	"hashbot/database"
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var enable = Command{
	Name:        "enable",
	Aliases:     []string{},
	Usage:       "enable [command]",
	Description: "Enables a command for all users in the channel",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdEnableDescription",
				Other: "Enables a command for all users in the channel",
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

		var (
			command Command
			ok      bool
			err     error
		)
		if command, ok = commandMap[args[1]]; !ok {
			found := false
			for _, cmd := range commandsNoPrefix {
				if cmd.Name == args[1] {
					command = cmd
					found = true
					break
				}
			}
			if !found {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID: "UnknownCommand",
					},
					TemplateData: map[string]string{
						"Command": args[1],
					},
				})
				sender.Say(message.Channel, msg)
				return nil
			}
		}

		if !command.CanDisable {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CommandCannotBeDisabled",
					Other: "❌Command '{{.Command}}' cannot be disabled",
				},
				TemplateData: map[string]string{
					"Command": args[1],
				},
			})

			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		if !(message.Chatter.IsMod || message.Chatter.IsBroadcaster) {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "MustBeModerator",
					Other: "❌You must be a moderator to use this command",
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		err = database.UpdateIsUserCommandEnabled(tx, true, message.RoomID, command.Name)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "EnabledCommands",
				Other: "✅Enabled command '{{.Command}}'",
			},
			TemplateData: map[string]string{
				"Command": command.Name,
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		return nil
	},
}
