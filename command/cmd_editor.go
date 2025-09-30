package command

import (
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var editor = Command{
	Name:        "editor",
	Aliases:     []string{},
	Usage:       "editor [add|remove] user",
	Description: "Add or remove 7TV emote editors",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdEditorDescription",
				Other: "Add or remove 7TV emote editors",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if parsedArgs.ArgCount != 2 {
			return false
		}
		if len(parsedArgs.Positional) != 2 {
			return false
		}
		if parsedArgs.Positional[0] != "add" && parsedArgs.Positional[0] != "remove" {
			return false
		}
		return true
	},
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if !message.Chatter.IsBroadcaster {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdEditorErrNotBroadcaster",
					Other: "❌Only the broadcaster can set editors",
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

		editor, err := twitchapi.GetUserByName(message.Cfg, args[2])
		if err != nil {
			return err
		}
		if len(editor) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "UserNotFound",
					Other: "User '{{.User}}' not found",
				},
				TemplateData: map[string]string{
					"User": args[2],
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		editorID, editorName := editor[0].ID, editor[0].Login
		var successMsg string
		switch args[1] {
		case "add":
			err = database.InsertEditor(tx, message.Chatter.ID, editorID, editorName)
			if err != nil {
				return err
			}

			successMsg = message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdEditorAdded",
					Other: "✅Added editor '{{.Editor}}'",
				},
				TemplateData: map[string]string{
					"Editor": args[2],
				},
			})
		case "remove":
			err = database.RemoveEditor(tx, message.Chatter.ID, editorID)
			if err != nil {
				return err
			}

			successMsg = message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdEditorRemoved",
					Other: "✅Removed editor '{{.Editor}}'",
				},
				TemplateData: map[string]string{
					"Editor": args[2],
				},
			})
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		sender.Say(message.Channel, successMsg, struct {
			Param types.SenderParam
			Value string
		}{Param: types.ReplyMessageID, Value: message.ID})
		return nil
	},
}
