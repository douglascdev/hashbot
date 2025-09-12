package command

import (
	"fmt"
	"monkebot/database"
	"monkebot/twitchapi"
	"monkebot/types"
)

var editor = types.Command{
	Name:              "editor",
	Aliases:           []string{},
	Usage:             "editor [add|remove] user",
	Description:       "Opt out of one or all commands",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) != 3 || args[1] != "add" && args[1] != "remove" {
			sender.Say(message.Channel, "Usage: editor [add|remove] user")
			return nil
		}

		if !message.Chatter.IsBroadcaster {
			sender.Say(message.Channel, "❌Only the broadcaster can set editors")
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
		if editor == nil || len(*editor) == 0 {
			sender.Say(message.Channel, "User not found", struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		editorID, editorName := (*editor)[0].ID, (*editor)[0].Login
		var successMsg string
		switch args[1] {
		case "add":
			err = database.InsertEditor(tx, message.Chatter.ID, editorID, editorName)
			if err != nil {
				return err
			}
			successMsg = fmt.Sprintf("✅Added editor %q", args[2])
		case "remove":
			err = database.RemoveEditor(tx, message.Chatter.ID, editorID)
			if err != nil {
				return err
			}
			successMsg = fmt.Sprintf("✅Removed editor %q", args[2])
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
