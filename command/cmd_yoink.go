package command

import (
	"errors"
	"fmt"
	"hashbot/database"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"hashbot/types"
	"strings"
)

var yoink = Command{
	Name:              "yoink",
	Aliases:           []string{},
	Usage:             "yoink [emote] #[channel] to:channel",
	Description:       "add given 7TV emote from a channel to another",
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

		var fromChannelName string
		if len(parsedArgs.HashPrefixed) == 1 {
			fromChannelName = parsedArgs.HashPrefixed[0]
		} else if name, found := parsedArgs.Named["from"]; found {
			fromChannelName = name
		} else {
			fromChannelName = message.Channel
		}

		var toChannelName string
		if name, found := parsedArgs.Named["to"]; found {
			toChannelName = name
		} else {
			toChannelName = message.Chatter.Name
		}

		channels, err := twitchapi.GetUserByName(message.Cfg, []string{fromChannelName, toChannelName}...)
		if err != nil || len(*channels) != 2 {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to fetch channel %q", fromChannelName))
			return err
		}
		fromChannel, toChannel := (*channels)[0], (*channels)[1]

		isBroadcaster := strings.EqualFold(toChannel.Login, message.Chatter.Name)
		if !isBroadcaster && !database.SelectIsEditor(tx, toChannel.ID, message.Chatter.ID) {
			sender.Say(message.Channel, "❌You must be an editor to use this command")
			return nil
		}

		fromChannelSTV, err := seventvapi.GetUserByConnection("https://7tv.io", fromChannel.ID)
		if err != nil {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to fetch sevenTV channel for %q", fromChannelName))
			return nil
		}

		emoteArgMap := make(map[string]bool)
		for _, emote := range parsedArgs.Positional {
			emoteArgMap[emote] = true
		}

		var emotes []struct{ ID, alias string }
		for _, set := range fromChannelSTV.Data.Users.UserByConnection.EmoteSets {
			if set.ID == fromChannelSTV.Data.Users.UserByConnection.Style.ActiveEmoteSetID {
				for _, emote := range set.Emotes.Items {
					if _, found := emoteArgMap[emote.Alias]; !found {
						continue
					}
					emotes = append(emotes, struct {
						ID    string
						alias string
					}{emote.ID, emote.Alias})
				}
				break
			}
		}

		if len(emotes) == 0 {
			sender.Say(message.Channel, fmt.Sprintf("❌Failed to find active emote set for %q", fromChannelName))
			return nil
		}

		for _, emote := range emotes {
			err = seventvapi.AddEmoteWithID("https://7tv.io", toChannel.ID, emote.ID, emote.alias, message.Cfg.SevenTVToken)
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
				return fmt.Errorf("failed to yoink emote: %w", err)
			}
		}

		return nil
	},
}
