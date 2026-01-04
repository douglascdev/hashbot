package command

import (
	"errors"
	"hashbot/game"
	"hashbot/types"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/rs/zerolog/log"
)

var duelMutex = sync.Mutex{}
var duels = []game.Duel{}

func findDuel(duels []game.Duel, name string) game.Duel {
	for _, duel := range duels {
		names := []string{duel.SourcePlayer().Name(), duel.TargetPlayer().Name()}
		if slices.Contains(names, strings.ToLower(name)) {
			return duel
		}
	}
	return nil
}

func removeDuel(duel game.Duel) {
	duelMutex.Lock()
	index := slices.Index(duels, duel)
	if index != -1 {
		duels = append(duels[:index], duels[index+1:]...)
		log.Info().Msg("removed duel between " + duel.SourcePlayer().Name() + " and " + duel.TargetPlayer().Name())
	}
	duelMutex.Unlock()
}

var gameCmd = Command{
	Name:              "game",
	Aliases:           []string{"jogo"},
	Usage:             "game duel [player] | game [potion | fireball | water | barrier]",
	Description:       "Duel with fatasswizardcat emotes",
	ChannelCooldown:   0,
	UserCooldown:      5,
	NoPrefix:          false,
	NoPrefixShouldRun: nil,
	CanDisable:        true,
	ValidUsage: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) bool {
		if len(parsedArgs.Positional) < 1 {
			return false
		}

		if !slices.Contains([]string{"duel", "potion", "fireball", "water", "barrier"}, parsedArgs.Positional[0]) {
			return false
		}

		return true
	},
	GetLocalizedDescription: func(localizer *i18n.Localizer) string {
		return localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "CmdGameDescription",
				Other: "Duel with fatasswizardcat emotes",
			},
		})
	},
	ExecuteParsed: func(message *types.Message, sender types.MessageSender, parsedArgs *ParseResult) error {
		tx, err := message.DB.Begin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		msgDuelNotExists := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "DuelNotExists",
				Other: "Found no active duel for player",
			},
		})

		switch parsedArgs.Positional[0] {
		case "duel":
			if len(parsedArgs.Positional) < 2 {
				msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "DuelNoTarget",
						Other: "Please set a duel target",
					},
				})
				return types.NewCommandError(errors.New("duel target not passed"), msg)
			}

			p1Name, p2Name := strings.ToLower(message.Chatter.Name), strings.ToLower(parsedArgs.Positional[1])
			for _, duel := range duels {
				names := []string{p1Name, p2Name}
				if slices.Contains(names, duel.SourcePlayer().Name()) || slices.Contains(names, duel.TargetPlayer().Name()) {
					msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "DuelExists",
							Other: "One of the players already has an active duel",
						},
					})
					return types.NewCommandError(errors.New("duel exists"), msg)
				}
			}

			p1, p2 := game.NewPlayer(p1Name), game.NewPlayer(p2Name)
			duel := game.NewDuel(p1, p2, message.Localizer)

			duelMutex.Lock()
			duels = append(duels, duel)
			duelMutex.Unlock()

			go func() {
				time.Sleep(time.Minute * 5)
				removeDuel(duel)
			}()

			next := ""
			if duel.NextTurnIsSource() {
				next = duel.SourcePlayer().Name()
			} else {
				next = duel.TargetPlayer().Name()
			}
			msg := message.Localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "DuelStart",
					Other: "duel started! Your turn {{.Player}}. Options are: fireball, water, potion",
				},
				TemplateData: map[string]string{
					"Player": next,
				},
			})
			sender.Say(message.Channel, msg)
			return nil

		case "fireball":
			duelFound := findDuel(duels, message.Chatter.Name)
			if duelFound == nil {
				return types.NewCommandError(errors.New("duel doesn't exist"), msgDuelNotExists)
			}
			result := duelFound.Do(strings.ToLower(message.Chatter.Name) == duelFound.SourcePlayer().Name(), "fireball")
			if duelFound.Winner() != nil {
				result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID:    "DuelWon",
						Other: "{{.Winner}} won! 👏",
					},
					TemplateData: map[string]string{
						"Winner": duelFound.Winner().Name(),
					},
				})
				removeDuel(duelFound)
			}
			sender.Say(message.Channel, result, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		case "potion":
			duelFound := findDuel(duels, message.Chatter.Name)
			if duelFound == nil {
				return types.NewCommandError(errors.New("duel doesn't exist"), msgDuelNotExists)
			}
			result := duelFound.Do(strings.ToLower(message.Chatter.Name) == duelFound.SourcePlayer().Name(), "potion")
			if duelFound.Winner() != nil {
				result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID: "DuelWon",
					},
					TemplateData: map[string]string{
						"Winner": duelFound.Winner().Name(),
					},
				})
				removeDuel(duelFound)
			}
			sender.Say(message.Channel, result, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		case "water":
			duelFound := findDuel(duels, message.Chatter.Name)
			if duelFound == nil {
				return types.NewCommandError(errors.New("duel doesn't exist"), msgDuelNotExists)
			}
			result := duelFound.Do(strings.ToLower(message.Chatter.Name) == duelFound.SourcePlayer().Name(), "water")
			if duelFound.Winner() != nil {
				result += message.Localizer.MustLocalize(&i18n.LocalizeConfig{
					DefaultMessage: &i18n.Message{
						ID: "DuelWon",
					},
					TemplateData: map[string]string{
						"Winner": duelFound.Winner().Name(),
					},
				})
				removeDuel(duelFound)
			}
			sender.Say(message.Channel, result, struct {
				Param types.SenderParam
				Value string
			}{types.ReplyMessageID, message.ID})
			return nil
		case "barrier":
		}

		return nil
	},
}
