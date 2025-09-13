package command

import (
	"errors"
	"fmt"
	"hashbot/database"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"hashbot/types"
)

var remove = Command{
	Name:              "remove",
	Aliases:           []string{"r"},
	Usage:             "remove [emote] #[channel]",
	Description:       "Removes given 7TV emote from the channel",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if parsedArgs.ArgCount == 1 {
			return false
		}
		return true
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}

		var targetChannelName string
		if len(parsedArgs.Prefixed) == 1 {
			targetChannelName = parsedArgs.Prefixed[0].Value
		} else {
			targetChannelName = message.Channel
		}

		res, err := twitchapi.GetUserByName(message.Cfg, targetChannelName)
		if err != nil || len(*res) == 0 {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to fetch channel %q", targetChannelName))
			return err
		}
		targetChannel := (*res)[0]

		if !message.Chatter.IsBroadcaster && !database.SelectIsEditor(tx, targetChannel.ID, message.Chatter.ID) {
			sender.Say(message.Channel, "❌You must be an editor to use this command")
			return nil
		}

		for _, arg := range parsedArgs.Positional {
			emote := arg.Value
			err = seventvapi.RemoveEmote("https://7tv.io", targetChannel.ID, emote, message.Cfg.SevenTVToken)
			if err != nil {
				if errors.Is(err, seventvapi.EmoteNotFound) {
					errorMsg := fmt.Sprintf("❌%s", err.Error())
					sender.Say(message.Channel, errorMsg, []struct {
						Param types.SenderParam
						Value string
					}{
						{types.ReplyMessageID, message.ID},
					}...)
					return nil
				}
				return fmt.Errorf("failed to remove emote: %w", err)
			}
		}

		return nil
	},
}
