package command

import (
	"fmt"
	"hashbot/types"
	"runtime/metrics"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var ping = Command{
	Name:        "ping",
	Aliases:     []string{},
	Usage:       "ping",
	Description: "Responds with pong and latency to twitch in milliseconds",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdPingDescription",
				Other: "Responds with pong and latency to twitch in milliseconds",
			},
		})
	},
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		responses := []string{
			"👋 glorp Pong!",
		}

		latency, err := sender.Ping()
		if err != nil {
			log.Warn().Err(err).Msg("failed to get latency for ping message")
		} else if latency == 0 {
			log.Warn().Msg("failed to get latency for ping message, client sent no pings yet")
		} else {
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "CmdPingLatency",
					Other: "Latency: %dms",
				},
			})
			responses = append(responses, fmt.Sprintf(msg, latency.Milliseconds()))
		}

		memSamples := []metrics.Sample{
			{Name: "/memory/classes/total:bytes"},
		}
		metrics.Read(memSamples)

		memoryMsg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdPingMemory",
				Other: "Memory: %d MiB",
			},
		})
		uptimeMsg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdPingUptime",
				Other: "Uptime: %s",
			},
		})

		responses = append(responses,
			fmt.Sprintf(memoryMsg, memSamples[0].Value.Uint64()/1024/1024),
			fmt.Sprintf(uptimeMsg, sender.Uptime().Round(time.Second)),
		)

		sender.Say(message.Channel, strings.Join(responses, " | "))
		return nil
	},
}
