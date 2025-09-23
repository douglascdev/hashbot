package command

import (
	"errors"
	"fmt"
	"hashbot/database"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"hashbot/types"
)

var add = Command{
	Name:              "add",
	Aliases:           []string{},
	Usage:             "add [emote] #[channel]",
	Description:       "adds given 7TV emote to the channel",
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
		if err != nil || len(*res) == 0 {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to fetch channel %q", targetChannelName))
			return err
		}
		targetChannel := (*res)[0]

		if !message.Chatter.IsBroadcaster && !database.SelectIsEditor(tx, targetChannel.ID, message.Chatter.ID) {
			sender.Say(message.Channel, "❌You must be an editor to use this command")
			return nil
		}

		targetStv, err := seventvapi.GetUserByConnection("https://7tv.io", targetChannel.ID)
		if err != nil {
			sender.Say(message.Channel, "❌Failed to fetch user's 7tv data")
			return nil
		}
		activeSetEmotes := make(map[string]bool)

		for _, set := range targetStv.Data.Users.UserByConnection.EmoteSets {
			if set.ID == targetStv.Data.Users.UserByConnection.Style.ActiveEmoteSetID {
				for _, emote := range set.Emotes.Items {
					activeSetEmotes[emote.Alias] = true
				}
			}
		}

		var (
			addedEmotes []string
			errorEmotes []string
		)
		for _, emote := range parsedArgs.Positional {
			err = seventvapi.AddEmoteWithQuery("https://7tv.io", targetChannel.ID, emote, message.Cfg.SevenTVToken)
			if _, found := activeSetEmotes[emote]; err != nil || found {
				errorEmotes = append(errorEmotes, emote)
				continue
			}
			addedEmotes = append(addedEmotes, emote)
		}

		var reply string
		if len(errorEmotes) > 0 {
			reply += fmt.Sprintf("Failed to add %d emotes: %q. ", len(errorEmotes), strings.Join(errorEmotes, ", "))
		}
		if len(addedEmotes) > 0 {
			reply += fmt.Sprintf("Added %d emotes: %q.", len(addedEmotes), strings.Join(addedEmotes, ", "))
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
