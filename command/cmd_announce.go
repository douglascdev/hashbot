package command

import (
	"hashbot/database"
	"hashbot/types"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var announce = Command{
	Name:              "announce",
	Aliases:           []string{},
	Usage:             "announce preview | announce",
	Description:       "(admin only) Make or preview pending announcements.",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdAnnounceDescription",
				Other: "(admin only) Make or preview pending announcements.",
			},
		})
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		isAdmin, _ := database.SelectIsUserAdmin(tx, message.Chatter.ID)
		if !isAdmin {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAnnounceAdmin",
					Other: "Only admins allowed...",
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		channels, err := database.SelectAnnounceNews(tx)
		if err != nil {
			log.Err(err).Msg("failed to announce news")
		}
		log.Info().Interface("channels", channels).Msg("inserting announced news")

		if len(channels) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAnnounceNoChannels",
					Other: "No channels left to announce in",
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		if len(parsedArgs.Positional) > 0 && parsedArgs.Positional[0] == "preview" {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdAnnouncePreview",
					Other: "Would announce to {{.Channels}} channels. Full list send to logs.",
				},
				TemplateData: map[string]int{
					"Channels": len(channels),
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		for _, channel := range channels {
			var msg string
			switch channel.ChannelLang {
			case "pt":
				msg = channel.PtTxt
			default:
				msg = channel.EnTxt
			}
			sender.Say(channel.ChannelName, msg)
			time.Sleep(time.Second / 4)
		}
		err = database.DeleteNews(tx)
		if err != nil {
			log.Err(err).Msg("failed to clear announced news")
		}
		err = tx.Commit()
		if err != nil {
			log.Err(err).Msg("failed to commit announced news")
		}

		return nil
	},
}
