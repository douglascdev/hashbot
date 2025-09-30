package command

import (
	"hashbot/types"
	"math/rand/v2"
	"regexp"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var buttRegexp = regexp.MustCompile(`^butt`)

var butt = Command{
	Name:        "butt",
	Aliases:     []string{},
	Usage:       "butt[anything]",
	Description: "Responds with butt to messages starting with butt",
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdButtDescription",
				Other: "Responds with butt to messages starting with butt",
			},
		})
	},
	ChannelCooldown: 5,
	UserCooldown:    5,
	NoPrefix:        true,
	NoPrefixShouldRun: func(message *types.Message, sender types.MessageSender, args []string) bool {
		return buttRegexp.MatchString(message.Message)
	},
	CanDisable: true,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {
		if rand.IntN(100) == 1 {
			sender.Say(message.Channel, "buttConcerned")
		} else {
			sender.Say(message.Channel, "butt")
		}
		return nil
	},
}
