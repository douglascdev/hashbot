package command

import (
	"errors"
	"fmt"
	"monkebot/database"
	"monkebot/seventvapi"
	"monkebot/twitchapi"
	"monkebot/types"
	"slices"
	"strings"
)

var remove = types.Command{
	Name:              "remove",
	Aliases:           []string{"r"},
	Usage:             "remove [emote] #[channel]",
	Description:       "Removes given 7TV emote from the channel",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if len(args) < 2 {
			sender.Say(message.Channel, "❌Usage: remove [emote] #[channel]")
			return nil
		}

		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}

		var targetChannelName string
		for i := 1; i < len(args); i++ {
			if channel, found := strings.CutPrefix(args[i], "#"); found {
				targetChannelName = channel
				// remove channel from args
				args = slices.Delete(args, i, i+1)
				break
			}
		}

		if targetChannelName == "" {
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

		//remove command name from args
		args = args[1:]

		for _, emote := range args {
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
