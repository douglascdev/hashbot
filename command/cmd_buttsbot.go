package command

import "hashbot/types"

var buttsbot = Command{
	Name:            "buttsbot",
	Aliases:         []string{},
	Usage:           "send any message in chat",
	Description:     "Replaces random syllables with butt",
	ChannelCooldown: 60,
	UserCooldown:    60,
	NoPrefix:        true,
	NoPrefixShouldRun: func(message *types.Message, sender types.MessageSender, args []string) bool {
		return sender.ShouldButtify()
	},
	CanDisable: true,
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		if len(parsedArgs.Links) > 0 {
			return nil
		}

		newSentence := sender.Buttify(message.Message)
		if newSentence != message.Message {
			sender.Say(message.Channel, newSentence)
		}

		return nil
	},
}
