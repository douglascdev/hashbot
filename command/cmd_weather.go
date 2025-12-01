package command

import (
	"fmt"
	"hashbot/database"
	"hashbot/openmeteoapi"
	"hashbot/twitchapi"
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var weather = Command{
	Name:              "weather",
	Aliases:           []string{"climate", "clima", "tempo"},
	Usage:             "weather | weather #user",
	Description:       "Get current weather for an user based on their defined location.",
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
				ID:    "CmdWeatherDescription",
				Other: "Get current weather for an user based on their defined location.",
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
					ID: "CmdTimeNotSet",
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

		lat, lon := fmt.Sprintf("%.2f", location.Latitude), fmt.Sprintf("%f", location.Longitude)
		weatherData, err := openmeteoapi.GetWeather(lat, lon)
		if err != nil {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdWeatherFailedToLoadLocation",
					Other: "Failed to load weather for '{{.Name}}'",
				},
				TemplateData: map[string]string{
					"Name": location.Name,
				},
			})
			sender.Say(message.Channel, msg, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		}

		var result string
		result += location.Name + ", "
		result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdWeatherApparentTemp",
				Other: " {{.CurrentTemp}}{{.CurrentTempUnits}}, feels like {{.Temp}}{{.Unit}}. ",
			},
			TemplateData: map[string]string{
				"CurrentTemp":      fmt.Sprintf("%.2f", weatherData.Current.Temperature2M),
				"CurrentTempUnits": weatherData.CurrentUnits.Temperature2M,
				"Temp":             fmt.Sprintf("%.2f", weatherData.Current.ApparentTemperature),
				"Unit":             weatherData.CurrentUnits.ApparentTemperature,
			},
		})
		if weatherData.Current.Rain == 0 {
			result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdWeatherNoRain",
					Other: "No rain. ",
				},
			})
		} else {
			result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdWeatherRain",
					Other: "Rain: {{.Rain}}{{.Unit}}. ",
				},
				TemplateData: map[string]string{
					"Rain": fmt.Sprintf("%.2f", weatherData.Current.Rain),
					"Unit": weatherData.CurrentUnits.Rain,
				},
			})
		}
		if weatherData.Current.Snowfall > 0 {
			result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdWeatherSnowfall",
					Other: "Snowfall: {{.Snowfall}}{{.Unit}}. ",
				},
				TemplateData: map[string]string{
					"Snowfall": fmt.Sprintf("%.2f", weatherData.Current.Snowfall),
					"Unit":     weatherData.CurrentUnits.Snowfall,
				},
			})
		}
		result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdWeatherWindspeed",
				Other: "Wind speed: {{.Speed}}{{.Unit}}. ",
			},
			TemplateData: map[string]string{
				"Speed": fmt.Sprintf("%.2f", weatherData.Current.WindSpeed10M),
				"Unit":  weatherData.CurrentUnits.WindSpeed10M,
			},
		})
		result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdWeatherCloudCover",
				Other: "Cloud cover: {{.Cover}}{{.Unit}}. ",
			},
			TemplateData: map[string]string{
				"Cover": fmt.Sprintf("%d", weatherData.Current.CloudCover),
				"Unit":  weatherData.CurrentUnits.CloudCover,
			},
		})
		sender.Say(message.Channel, result, struct {
			Param types.SenderParam
			Value string
		}{types.ReplyMessageID, message.ID})
		return nil
	},
}
