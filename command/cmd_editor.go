package command

import (
	"fmt"
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
)

var editor = Command{
	Name:              "editor",
	Aliases:           []string{},
	Usage:             "editor [add|remove] user",
	Description:       "Add or remove 7TV emote editors",
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
		if editor == nil || len(editor) == 0 {
			sender.Say(message.Channel, "User not found", struct {
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
