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

var add = Command{
	Name:              "add",
	Aliases:           []string{},
	Usage:             "add [emote] #[channel]",
	Description:       "Adds given 7TV emote to the channel",
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
					ID:    "CmdAddFailedNotEditor",
					Other: "❌You must be an editor to use this command",
				},
			})
			sender.Say(message.Channel, msg)
			return nil
		}

		targetStv, err := seventvapi.GetUserByConnection("https://7tv.io", targetChannel.ID)
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAddFailedToFetch7TV",
					Other: "❌Failed to fetch user's 7tv data",
				},
			})
			sender.Say(message.Channel, msg)
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
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAddFailedToAddEmote",
					One:   "Failed to add {{.Count}} emote: {{.Emotes}}",
					Other: "Failed to add {{.Count}} emotes: {{.Emotes}}",
				},
				PluralCount: len(errorEmotes),
				TemplateData: map[string]any{
					"Count":  len(errorEmotes),
					"Emotes": strings.Join(errorEmotes, " "),
				},
			})
			reply += msg
		}
		if len(addedEmotes) > 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAddAddedEmote",
					One:   "Added {{.Count}} emote: {{.Emotes}}",
					Other: "Added {{.Count}} emotes: {{.Emotes}}",
				},
				PluralCount: len(addedEmotes),
				TemplateData: map[string]any{
					"Count":  len(addedEmotes),
					"Emotes": strings.Join(addedEmotes, " "),
				},
			})
			reply += msg
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
