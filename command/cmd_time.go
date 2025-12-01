package command

import (
	"hashbot/database"
	"hashbot/twitchapi"
	"hashbot/types"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var cmdTime = Command{
	Name:              "time",
	Aliases:           []string{"hora", "horario"},
	Usage:             "time | time #user",
	Description:       "Get user's time based on their defined location.",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		return len(parsedArgs.Positional) == 0
	},
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdtimeDescription",
				Other: "Get user's time based on their defined location.",
			},
		})
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}
		// no args = default to get sender's time
		var targetUser = message.Chatter.Name

		if len(parsedArgs.HashPrefixed) > 0 {
			targetUser = parsedArgs.HashPrefixed[0]
		}

		targetHelixUsers, err := twitchapi.GetUserByName(message.Cfg, targetUser)
		if err != nil || len(targetHelixUsers) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdTimeTargetNotFound",
					Other: "User {{.User}} not found.",
				},
				TemplateData: map[string]string{
					"User": targetUser,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		location, err := database.SelectUserLocation(tx, targetHelixUsers[0].ID)
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdTimeNotSet",
					Other: "Location not set for {{.User}}.",
				},
				TemplateData: map[string]string{
					"User": targetUser,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		timeLocation, err := time.LoadLocation(location.Timezone)
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdTimeFailedToLoadLocation",
					Other: "Failed to load {{.User}}'s timezone '{{.Timezone}}'",
				},
				TemplateData: map[string]string{
					"User":     targetUser,
					"Timezone": location.Timezone,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdTimeResult",
				Other: "User {{.User}}'s time is '{{.Time}}'",
			},
			TemplateData: map[string]string{
				"User": targetUser,
				"Time": time.Now().In(timeLocation).Format("15:04"),
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})
		return nil
	},
}
