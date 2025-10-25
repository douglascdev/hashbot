package command

import (
	"errors"
	"fmt"
	"hashbot/database"
	"hashbot/seventvapi"
	"hashbot/twitchapi"
	"hashbot/types"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var yoink = Command{
	Name:        "yoink",
	Aliases:     []string{},
	Usage:       "yoink [emote] #[channel] to:channel",
	Description: "Add given 7TV emote from a channel to another",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdYoinkDescription",
				Other: "Add given 7TV emote from a channel to another",
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

		if strings.EqualFold(fromChannelName, toChannelName) {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "YoinkFromChannelToChannel",
					Other: "❌Can't yoink from '{{.From}}' to '{{.To}}', they are equal.",
				},
				TemplateData: map[string]string{
					"From": fromChannelName,
					"To":   toChannelName,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return err
		}

		channels, err := twitchapi.GetUserByName(message.Cfg, fromChannelName, toChannelName)
		log.Debug().Interface("channels", channels).Msg("fetched channels")
		if err != nil || len(channels) != 2 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FailedToFetchChannel",
					Other: "❌Failed to fetch channel '{{.Channel}}'",
				},
				TemplateData: map[string]string{
					"Channel": fromChannelName,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return err
		}
		fromChannel, toChannel := channels[0], channels[1]

		isBroadcaster := strings.EqualFold(toChannel.Login, message.Chatter.Name)
		if !isBroadcaster && !database.SelectIsEditor(tx, toChannel.ID, message.Chatter.ID) {
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

		fromCh := make(chan *seventvapi.GetUserByConnectionResp)
		fromChErr := make(chan error)
		toCh := make(chan *seventvapi.GetUserByConnectionResp)
		toChErr := make(chan error)
		go func() {
			fromChannelSTV, err := seventvapi.GetUserByConnection("https://7tv.io", fromChannel.ID, message.Cfg.SevenTVToken)
			fromCh <- fromChannelSTV
			fromChErr <- err
		}()
		go func() {
			toChannelSTV, err := seventvapi.GetUserByConnection("https://7tv.io", fromChannel.ID, message.Cfg.SevenTVToken)
			toCh <- toChannelSTV
			toChErr <- err
		}()
		fromChannelSTV, err := <-fromCh, <-fromChErr
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FailedFetchSevenTVChannel",
					Other: "Failed to fetch sevenTV channel for '{{.Channel}}'",
				},
				TemplateData: map[string]string{
					"Channel": fromChannelName,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		toChannelSTV, err := <-toCh, <-toChErr
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FailedFetchSevenTVChannel",
					Other: "Failed to fetch sevenTV channel for '{{.Channel}}'",
				},
				TemplateData: map[string]string{
					"Channel": fromChannelName,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		emoteFoundInSet := make(map[string]bool)
		for _, emote := range parsedArgs.Positional {
			emoteFoundInSet[emote] = false
		}

		var emoteAlreadyInFromCh []string
		for _, set := range toChannelSTV.Data.Users.UserByConnection.EmoteSets {
			if set.ID == toChannelSTV.Data.Users.UserByConnection.Style.ActiveEmoteSetID {
				for _, emote := range set.Emotes.Items {
					if _, found := emoteFoundInSet[emote.Alias]; found {
						emoteAlreadyInFromCh = append(emoteAlreadyInFromCh, emote.Alias)
					}
				}
			}
		}

		if len(emoteAlreadyInFromCh) > 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "EmotesAlreadyInChannel",
					One:   "❌Emote '{{.Emotes}}' already in '{{.ToChannel}}'.",
					Other: "❌Emotes '{{.Emotes}}' already in '{{.ToChannel}}'.",
				},
				PluralCount: len(emoteAlreadyInFromCh),
				TemplateData: map[string]string{
					"ToChannel": toChannelName,
					"Emotes":    strings.Join(emoteAlreadyInFromCh, ", "),
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		var emotes []struct{ ID, alias string }
		log.Debug().Int("lenEmoteSets", len(fromChannelSTV.Data.Users.UserByConnection.EmoteSets)).Msg("")
		for _, set := range fromChannelSTV.Data.Users.UserByConnection.EmoteSets {
			log.Debug().Str("setID", set.ID).Str("activeSetID", fromChannelSTV.Data.Users.UserByConnection.Style.ActiveEmoteSetID).Msg("finding active set")
			if set.ID == fromChannelSTV.Data.Users.UserByConnection.Style.ActiveEmoteSetID {
				for _, emote := range set.Emotes.Items {
					if _, found := emoteFoundInSet[emote.Alias]; !found {
						continue
					}
					emotes = append(emotes, struct {
						ID    string
						alias string
					}{emote.ID, emote.Alias})
					emoteFoundInSet[emote.Alias] = true
				}
				break
			}
		}

		var notFoundEmotes []string
		for emote, found := range emoteFoundInSet {
			if !found {
				notFoundEmotes = append(notFoundEmotes, emote)
			}
		}

		if len(notFoundEmotes) > 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FailedFindEmote",
					One:   "❌Failed to find emote '{{.Emotes}}' in '{{.Channel}}'.",
					Other: "❌Failed to find emotes '{{.Emotes}}' in '{{.Channel}}'.",
				},
				PluralCount: len(notFoundEmotes),
				TemplateData: map[string]string{
					"Channel": fromChannelName,
					"Emotes":  strings.Join(notFoundEmotes, ", "),
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		if len(emotes) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "FailedFindActiveEmoteSet",
					Other: "❌Failed to find active emote set for {{.Channel}}",
				},
				TemplateData: map[string]string{
					"Channel": fromChannelName,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{Param: types.ReplyMessageID, Value: message.ID})
			return nil
		}

		var added []string
		for _, emote := range emotes {
			err = seventvapi.AddEmoteWithID("https://7tv.io", toChannel.ID, emote.ID, emote.alias, message.Cfg.SevenTVToken)
			if err != nil {
				if errors.Is(err, seventvapi.EmoteNotFound) {
					msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "EmoteNotFound",
							Other: "❌Emote '{{.Emote}}' not found",
						},
					})
					sender.Say(message.Channel, msg, []struct {
						Param types.SenderParam
						Value string
					}{
						{types.ReplyMessageID, message.ID},
					}...)
					return nil
				}
				return fmt.Errorf("failed to yoink emote: %w", err)
			}

			added = append(added, emote.alias)
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "AddedEmotes",
				One:   "✅Added emote '{{.Emotes}}' from '{{.FromChannel}}' to '{{.ToChannel}}'",
				Other: "✅Added emotes '{{.Emotes}}' from '{{.FromChannel}}' to '{{.Channel}}'",
			},
			PluralCount: len(added),
			TemplateData: map[string]string{
				"Emotes":      strings.Join(added, ", "),
				"FromChannel": fromChannelName,
				"ToChannel":   toChannelName,
			},
		})

		sender.Say(message.Channel, msg, []struct {
			Param types.SenderParam
			Value string
		}{
			{types.ReplyMessageID, message.ID},
		}...)

		return nil
	},
}
