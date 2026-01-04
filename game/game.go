package game

import (
	"errors"
	"fmt"
	"math/rand/v2"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Player interface {
	Name() string
	HP() int
	Barrier() int
	UsedPotions() int

	SetHP(int)
	SetBarrier(int)
	IncrementUsedPotions() error
}

type player struct {
	name        string
	hp          int
	barrier     int
	usedPotions int
}

func (p *player) Name() string {
	return p.name
}
func (p *player) HP() int {
	return p.hp
}
func (p *player) Barrier() int {
	return p.barrier
}

func (p *player) UsedPotions() int {
	return p.usedPotions
}

func (p *player) SetHP(hp int) {
	p.hp = hp
}

func (p *player) SetBarrier(b int) {
	p.barrier = b
}

func (p *player) IncrementUsedPotions() error {
	if p.usedPotions == 1 {
		return errors.New("can't use more than 1 potion")
	}
	p.usedPotions += 1
	return nil
}

func NewPlayer(name string) Player {
	return &player{name: name, hp: 100}
}

type Action interface {
	Emote() string
	Apply(source Player, target Player, localizer *i18n.Localizer) string
}

type damageRange struct {
	start int
	end   int
}

type action struct {
	emote             string
	targetDamageRange *damageRange
	sourceBarrier     int
	sourceHeal        int
}

func (a *action) Emote() string {
	return a.emote
}

func (a *action) Apply(source Player, target Player, localizer *i18n.Localizer) string {
	effect := ""

	if a.sourceHeal != 0 {
		err := source.IncrementUsedPotions()
		if err != nil {
			effect += localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "IncrementPotionErr",
					Other: "{{.Source}}, you can't use more than 1 potion!",
				},
				TemplateData: map[string]string{
					"Source": source.Name(),
				},
			})

		} else {
			source.SetHP(source.HP() + a.sourceHeal)
			effect += localizer.MustLocalize(&i18n.LocalizeConfig{
				DefaultMessage: &i18n.Message{
					ID:    "SetHPEffect",
					Other: "{{.Source}} healed => {{.Action}} {{.HP}}🧪",
				},
				TemplateData: map[string]string{
					"Action": a.emote,
					"HP":     fmt.Sprintf("%d", a.sourceHeal),
					"Source": source.Name(),
				},
			})
		}
	}

	if a.sourceBarrier != 0 {
		source.SetBarrier(source.Barrier() + a.sourceBarrier)
		effect += localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SetBarrierEffect",
				Other: "{{.Source}} applied a barrier => {{.Action}} {{.Barrier}}🛡️",
			},
			TemplateData: map[string]string{
				"Action":  a.emote,
				"Barrier": fmt.Sprintf("%d", a.sourceBarrier),
				"Source":  source.Name(),
			},
		})
	}

	if a.targetDamageRange != nil {
		damage := rand.IntN(a.targetDamageRange.end-a.targetDamageRange.start) + a.targetDamageRange.start
		effect += localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "SetDamageEffect",
				Other: "{{.Source}} attacked {{.Target}} => {{.Action}} {{.Health}}❤️ - {{.Damage}}❤️ + {{.Barrier}}🛡️ = {{.HpLeft}}❤️",
			},
			TemplateData: map[string]string{
				"Source":  source.Name(),
				"Target":  target.Name(),
				"Action":  a.emote,
				"Health":  fmt.Sprintf("%d", target.HP()),
				"Damage":  fmt.Sprintf("%d", damage),
				"Barrier": fmt.Sprintf("%d", target.Barrier()),
				"HpLeft":  fmt.Sprintf("%d", target.HP()-damage),
			},
		})
		target.SetHP(target.HP() - damage + target.Barrier())
	}

	return effect
}

func NewAction(actionName string, localizer *i18n.Localizer) (Action, error) {
	switch actionName {
	case "potion":
		return &action{emote: "fatasswizardcatdrunk", sourceHeal: 50}, nil
	case "barrier":
		return &action{emote: "fatasswizardcastsbarrier", sourceBarrier: 10}, nil
	case "fireball":
		return &action{emote: "fatasswizardcastsafireballonyou", targetDamageRange: &damageRange{start: 30, end: 50}}, nil
	case "water":
		return &action{emote: "fatasswizardcastswater", targetDamageRange: &damageRange{start: 35, end: 40}}, nil
	default:
		msg := localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ActionNotFound",
				Other: "action '{{.Action}}' not found",
			},
			TemplateData: map[string]string{
				"Action": actionName,
			},
		})

		return nil, errors.New(msg)
	}
}

type Duel interface {
	SourcePlayer() Player
	TargetPlayer() Player
	Localizer() *i18n.Localizer
	Do(doneBySource bool, actionName string) string
	Winner() Player
	Actions() []Action
	NextTurnIsSource() bool
}

type duel struct {
	source           Player
	target           Player
	nextTurnIsSource bool
	localizer        *i18n.Localizer
	actions          []Action
}

func (d *duel) Localizer() *i18n.Localizer {
	return d.localizer
}

func (d *duel) SourcePlayer() Player {
	return d.source
}

func (d *duel) TargetPlayer() Player {
	return d.target
}

func (d *duel) Do(doneBySource bool, actionName string) string {
	if doneBySource && !d.nextTurnIsSource {
		return d.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "WrongTurn",
				Other: "the next play is {{.Target}}'s, not yours!",
			},
			TemplateData: map[string]string{
				"Target": d.target.Name(),
			},
		})
	} else if !doneBySource && d.nextTurnIsSource {
		return d.localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "WrongTurn",
				Other: "the next play is {{.Target}}'s, not yours!",
			},
			TemplateData: map[string]string{
				"Target": d.source.Name(),
			},
		})
	}
	action, err := NewAction(actionName, d.localizer)
	d.nextTurnIsSource = !d.nextTurnIsSource
	if err != nil {
		return err.Error()
	}
	d.actions = append(d.actions, action)
	if doneBySource {
		return action.Apply(d.source, d.target, d.localizer)
	} else {
		return action.Apply(d.target, d.source, d.localizer)
	}
}

func (d *duel) Winner() Player {
	if d.source.HP() <= 0 {
		return d.target
	}

	if d.target.HP() <= 0 {
		return d.source
	}

	return nil
}

func (d *duel) Actions() []Action {
	return d.actions
}

func (d *duel) NextTurnIsSource() bool {
	return d.nextTurnIsSource
}

func NewDuel(source Player, target Player, localizer *i18n.Localizer) Duel {
	firstTurnBySource := rand.IntN(2) == 0
	return &duel{localizer: localizer, source: source, target: target, actions: []Action{}, nextTurnIsSource: firstTurnBySource}
}
