package command

import (
	"hashbot/types"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

var help = Command{
	Name:              "help",
	Aliases:           []string{"commands"},
	Usage:             "help | help [command]",
	Description:       "Get the full list of commands, or help with a specific command",
	ChannelCooldown:   5,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        false,
	Execute: func(message *types.Message, sender types.MessageSender, args []string) error {

		msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "Help",
				Other: "Commands: http://hashbot.dev ● For help with a specific command: help <command>",
			},
		})
		if len(args) <= 1 {
			sender.Say(message.Channel, msg)
			return nil
		}

		var (
			command Command
			ok      bool
		)
		if command, ok = commandMap[args[1]]; !ok {
			found := false
			for _, cmd := range commandsNoPrefix {
				if cmd.Name == args[1] {
					command = cmd
					found = true
					break
				}
			}
			if !found {

				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "HelpUnknownCommand",
						Other: "❌Unknown command '{{.Command}}'.",
					},
					TemplateData: map[string]string{
						"Command": args[1],
					},
				})
				sender.Say(message.Channel, msg)
				return nil
			}
		}

		msg = message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "HelpUsage",
				Other: "Usage: {{.Usage}}",
			},
			TemplateData: map[string]string{
				"Usage": command.Usage,
			},
		})
		sender.Say(message.Channel, msg)
		return nil
	},
}
