package command

import (
	"hashbot/database"
	"hashbot/types"

	"github.com/douglascdev/buttifier"
)

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

		var (
			b   *buttifier.Buttifier
			err error
		)
		switch message.Lang {
		case "pt":
			b, err = buttifier.New(buttifier.Portuguese)
		default:
			b, err = buttifier.New(buttifier.English)
		}

		if err != nil {
			return err
		}

		tx, err := message.DB.Begin()
		if err != nil {
			return err
		}
		buttWord, err := database.SelectUserButtword(tx, message.RoomID)
		if err != nil {
			return err
		}

		b.ButtWord = buttWord

		newSentence := b.ButtifySentence(message.Message)
		if newSentence != message.Message {
			sender.Say(message.Channel, newSentence)
		}

		return nil
	},
}
