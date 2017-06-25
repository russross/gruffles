package main

import "time"

type MobController interface {
	SendMessage(string)
}

const (
	TimeToMove time.Duration = 200 * time.Millisecond
)

type pronouns int

const (
	PronounsIt pronouns = iota
	PronounsShe
	PronounsHe
)

type state int

const (
	StateStanding state = iota
	StateSleeping
	StateResting
	StateSitting
	StateFighting
	StateDead
	StateZombie
	StateFightingZombie
)

type Roll struct {
	Dice  int
	Faces int
	Add   int
}

type controller string

type Mob struct {
	Name        string
	Description string
	Title       string

	Location      *Room
	StartLocation *Room
	Visited       []bool

	Skills    []*Skill
	Effects   []*Effect
	Inventory []*Item
	Equipped  []*Item
	Pops      []Pop

	State                      state
	HP, HPNatural, HPMax       int
	Mana, ManaNatural, ManaMax int
	Move, MoveNatural, MoveMax int
	Str, StrNatural            int
	Con, ConNatural            int
	Dex, DexNatural            int
	Int, IntNatural            int
	Wis, WisNatural            int
	Pronouns                   pronouns
	Hit                        Roll
	Damage                     Roll
	Dodge                      Roll
	Absorb                     Roll
	Alignment                  int
	Level                      int
	Experience                 int
	Gold                       int

	SlowQueue        []string
	SlowPending      bool
	SlowBlockedUntil time.Time
	FastQueue        []string
	FastPending      bool
	FastBlockedUntil time.Time

	// When in a fight
	Opponent *Mob

	// Controller info
	Controller *MobController
	Player     *Player
}

func (mob *Mob) Send(msgType MsgType, msg string) {
	if mob.Player != nil {
		mob.Player.Send(Msg{
			Type:    msgType,
			Message: msg,
		})
	} else if mob.Controller != nil {
		// TODO: send message to non-player mob
	}
}
