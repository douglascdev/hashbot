package game

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// TODO: random status that if accepted also gives a penalty

type ItemType int

const (
	Consumable ItemType = iota
	Weapon
)

type Item struct {
	Name   string
	Emoji  string
	Amount int
	ItemType

	// combat stats
	Damage        int
	Armor         int
	SpellDamage   int
	BlockStrength int
}

var ItemNotFoundErr = errors.New("item not found")
var RaceNotFoundErr = errors.New("race not found")
var ExplorationTargetNotFoundErr = errors.New("exploration target not found")

func NewItem(name string, localizer *i18n.Localizer) (*Item, error) {
	pill := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ItemPill",
			Other: "Pill",
		},
	})

	dagger := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ItemDagger",
			Other: "Dagger",
		},
	})

	wand := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ItemWand",
			Other: "Wand",
		},
	})
	switch name {
	case pill:
		return &Item{Name: pill, Emoji: "💊", Amount: 1, ItemType: Consumable}, nil
	case dagger:
		return &Item{Name: dagger, Emoji: "🗡️", Amount: 1, ItemType: Weapon, Damage: 5, BlockStrength: 2}, nil
	case wand:
		return &Item{Name: wand, Emoji: "🪄", Amount: 1, ItemType: Weapon, Damage: 1, SpellDamage: 6, BlockStrength: 1}, nil
	default:
		return nil, ItemNotFoundErr
	}
}

type AttackMove struct {
	Name string

	// combat
	BaseDamage        int
	CritChance        int
	CritDmgMultiplier int

	MinimumLevel int
	Race
}

type Stats struct {
	Strength     int
	Intelligence int
	Dexterity    int
	Luck         int
}

type Race struct {
	Name        string
	ExtraStatus Stats
}

func NewRace(localizer *i18n.Localizer, race string) (*Race, error) {
	cat := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RaceCat",
			Other: "Cat",
		},
	})
	wolf := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RaceWolf",
			Other: "Wolf",
		},
	})
	owl := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RaceOwl",
			Other: "Wolf",
		},
	})

	switch race {
	case cat:
		return &Race{Name: cat, ExtraStatus: Stats{Strength: 2, Intelligence: 1, Dexterity: 5, Luck: 2}}, nil
	case wolf:
		return &Race{Name: cat, ExtraStatus: Stats{Strength: 5, Intelligence: 1, Dexterity: 3, Luck: 1}}, nil
	case owl:
		return &Race{Name: cat, ExtraStatus: Stats{Strength: 2, Intelligence: 1, Dexterity: 5, Luck: 2}}, nil
	default:
		return nil, RaceNotFoundErr
	}
}

type Player struct {
	Name      string
	Exp       int
	Health    int
	Inventory map[string]Item

	Stats
	Race
}

func (p *Player) Level() int {
	return p.Exp % 10
}

func (p *Player) ShowInventory(localizer *i18n.Localizer) string {
	var items []string
	for _, item := range p.Inventory {
		items = append(items, fmt.Sprintf("%s %sx%d", item.Name, item.Emoji, item.Amount))
	}
	return "[" + strings.Join(items, " | ") + "]"
}

func (p *Player) calcStats() Stats {
	return Stats{
		Strength:     p.Strength + p.Race.ExtraStatus.Strength,
		Intelligence: p.Intelligence + p.ExtraStatus.Intelligence,
		Dexterity:    p.Dexterity + p.Race.ExtraStatus.Dexterity,
		Luck:         p.Luck + p.Race.ExtraStatus.Luck,
	}
}

type CombatPower struct {
	PhysicalPower int
	MagicPower    int
	AttackSpeed   int
}

func (p *Player) CombatPower() *CombatPower {
	s := p.calcStats()
	return &CombatPower{
		PhysicalPower: s.Strength + s.Dexterity/2,
		MagicPower:    s.Intelligence + s.Luck/2,
		AttackSpeed:   1 / s.Dexterity,
	}
}

func (p *Player) ShowStats() string {
	s := p.calcStats()
	return fmt.Sprintf("%s{str: %d int: %d dex: %d lck: %d}", p.Name, s.Strength, s.Intelligence, s.Dexterity, s.Luck)
}

type ExplorationTarget struct {
	SuccessText     string
	FailText        string
	MinGainedExp    int
	MaxGainedExp    int
	GainedItems     []Item
	DifficultyLevel int
}

func (e *ExplorationTarget) NewExplorationTarget(localizer *i18n.Localizer, target string) (*ExplorationTarget, error) {
	mordor := localizer.MustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "PlaceMordor",
			Other: "Mordor",
		},
	})
	switch target {
	case mordor:
		return &ExplorationTarget{SuccessText: localizer.MustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "PlaceMordorSuccess",
				Other: "",
			},
		})}, nil
	}
	return nil, nil
}

func (e *ExplorationTarget) DidSucceed(p *Player) bool {
	if p.Level() < e.DifficultyLevel {
		return false
	}
	return true
}

/*
;g new name:buh race:cat
Criado gato "buh", digite ';g explore' ou ';g duel' ou ';g duel #player'

;g inv
[Pill 💊x3]

;g explore
Localizacoes disponiveis:
*/
