package command

import (
	"fmt"
	"hashbot/database"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"hashbot/types"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var remove = Command{
	Name:        "remove",
	Aliases:     []string{"r"},
	Usage:       "remove [emote] #[channel]",
	Description: "Removes given 7TV emote from the channel",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdRemoveDescription",
				Other: "Removes given 7TV emote from the channel",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) == 0 {
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

		var targetChannelName string
		if len(parsedArgs.HashPrefixed) == 1 {
			targetChannelName = parsedArgs.HashPrefixed[0]
		} else {
			targetChannelName = message.Channel
		}

		res, err := twitchapi.GetUserByName(message.Cfg, targetChannelName)
		if err != nil || len(res) == 0 {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to fetch channel %q", targetChannelName))
			return err
		}
		targetChannel := res[0]

		if !message.Chatter.IsBroadcaster && !database.SelectIsEditor(tx, targetChannel.ID, message.Chatter.ID) {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "NotEditor",
					Other: "❌You must be an editor to use this command",
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		var (
			removedEmotes []string
			errorEmotes   []string
		)
		for _, emote := range parsedArgs.Positional {
			err = seventvapi.RemoveEmote("https://7tv.io", targetChannel.ID, emote, message.Cfg.SevenTVToken)
			if err != nil {
				errorEmotes = append(errorEmotes, emote)
				continue
			}
			removedEmotes = append(removedEmotes, emote)
		}

		var reply string
		if len(errorEmotes) > 0 {
			reply += fmt.Sprintf("Failed to remove %d emote(s): %s ", len(errorEmotes), strings.Join(errorEmotes, " "))
		}
		if len(removedEmotes) > 0 {
			reply += fmt.Sprintf("Removed %d emote(s): %s ", len(removedEmotes), strings.Join(removedEmotes, " "))
		}
		sender.Say(message.Channel, reply, []struct {
			Param types.SenderParam
			Value string
		}{
			{types.ReplyMessageID, message.ID},
		}...)

		return nil
	},
}
