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

	SetHP(int)
	SetBarrier(int)
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
				Other: "{{.Source}} attacked {{.Target}} => {{.Action}} {{.Health}}❤️ - {{.Damage}}❤️ = {{.HpLeft}}❤️",
			},
			TemplateData: map[string]string{
				"Source": source.Name(),
				"Target": target.Name(),
				"Action": a.emote,
				"Health": fmt.Sprintf("%d", target.HP()),
				"Damage": fmt.Sprintf("%d", damage),
				"HpLeft": fmt.Sprintf("%d", target.HP()-damage),
			},
		})
		target.SetHP(target.HP() - damage)
	}

	return effect
}

func NewAction(actionName string, localizer *i18n.Localizer) (Action, error) {
	switch actionName {
	case "potion":
		return &action{emote: "fatasswizardcatdrunk", sourceHeal: 20}, nil
	case "barrier":
		return &action{emote: "fatasswizardcastsbarrier", sourceBarrier: 30}, nil
	case "fireball":
		return &action{emote: "fatasswizardcastsafireballonyou", targetDamageRange: &damageRange{start: 20, end: 40}}, nil
	case "water":
		return &action{emote: "fatasswizardcastswater", targetDamageRange: &damageRange{start: 25, end: 30}}, nil
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
	}
	action, err := NewAction(actionName, d.localizer)
	if err != nil {
		return err.Error()
	}
	d.actions = append(d.actions, action)
	d.nextTurnIsSource = !d.nextTurnIsSource
	return action.Apply(d.source, d.target, d.localizer)
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

func NewDuel(source Player, target Player, localizer *i18n.Localizer) Duel {
	firstTurnBySource := rand.IntN(2) == 0
	return &duel{localizer: localizer, source: source, target: target, actions: []Action{}, nextTurnIsSource: firstTurnBySource}
}
