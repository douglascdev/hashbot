package command

import (
	"hashbot/database"
	"hashbot/openmeteoapi"
	"hashbot/twitchapi"
	"hashbot/types"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var location = Command{
	Name:              "location",
	Aliases:           []string{"localizacao", "localização"},
	Usage:             "location [location]",
	Description:       "Set or get user's location. Used in commands related to weather and time.",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdLocationDescription",
				Other: "Set or get user's location. Used in commands related to weather and time.",
			},
		})
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}
		// no args = get user's location instead of setting
		var targetUser = message.Chatter.Name

		if len(parsedArgs.HashPrefixed) > 0 {
			targetUser = parsedArgs.HashPrefixed[0]
		}

		targetHelixUsers, err := twitchapi.GetUserByName(message.Cfg, targetUser)
		if err != nil || len(targetHelixUsers) == 0 {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdLocationTargetNotFound",
					Other: "User '{{.User}}' not found.",
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

		// no args = get user location
		if len(parsedArgs.Positional) == 0 {
			location, err := database.SelectUserLocation(tx, targetHelixUsers[0].ID)
			if err != nil {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "CmdLocationNotSet",
						Other: "Location not set for user '{{.User}}'.",
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

			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdLocationGet",
					Other: "User {{.User}}'s location is '{{.Location}}'",
				},
				TemplateData: map[string]string{
					"User":     targetUser,
					"Location": location.Name,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		isAdmin, _ := database.SelectIsUserAdmin(tx, message.Chatter.ID)
		if targetUser != message.Chatter.Name && !isAdmin {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdLocationDisallowedSetLang",
					Other: "Not allowed to set language for '{{.User}}'.",
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

		query := strings.Join(parsedArgs.Positional, "%20")
		locations, err := openmeteoapi.FindLocation(query)
		if err != nil || (locations != nil && len(locations.Results) == 0) {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdLocationNotFound",
					Other: "Location '{{.Location}}' not found.",
				},
				TemplateData: map[string]string{
					"Location": query,
				},
			})
			return types.NewCommandError(err, msg)
		}

		location := locations.Results[0]

		// use the channel's user language as the defult for the new user
		channelUserLanguage, err := database.SelectUserLanguage(tx, message.RoomID)
		if err != nil {
			log.Err(err).Str("channel", message.Channel).Str("user", message.Chatter.Name).Msg("failed to select channel's language")
		}

		err = database.InsertUsers(tx, false, channelUserLanguage, struct{ ID, Name string }{targetHelixUsers[0].ID, targetHelixUsers[0].Login})
		if err != nil {
			return err
		}

		if err = database.UpdateUserLocation(tx, targetHelixUsers[0].ID, &location); err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdLocationFailedToSet",
					Other: "Failed to set {{.User}}'s location  to '{{.Location}}'.",
				},
				TemplateData: map[string]string{
					"Location": location.Name,
					"User":     targetUser,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdLocationSuccess",
				Other: "✅Location for '{{.User}}' set to '{{.Location}}'.",
			},
			TemplateData: map[string]string{
				"Location": location.Name,
				"User":     targetUser,
			},
		})
		sender.Say(message.Channel, msg, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})

		return nil
	},
}
