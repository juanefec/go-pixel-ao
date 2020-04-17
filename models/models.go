package models

import "github.com/segmentio/ksuid"

type PlayerMessage struct {
	Spell  SpellMsg  `json:"spell"`
	Player PlayerMsg `json:"player"`
}

type UpdateMessage struct {
	Spells  []*SpellMsg  `json:"spells"`
	Players []*PlayerMsg `json:"players"`
}
type SpellMsg struct {
	ID   ksuid.KSUID `json:"id"`
	Name string      `json:"name"`
	X    float64     `json:"x"`
	Y    float64     `json:"y"`
}
type PlayerMsg struct {
	ID     ksuid.KSUID `json:"id"`
	Name   string      `json:"name"`
	X      float64     `json:"x"`
	Y      float64     `json:"y"`
	Dir    string      `json:"dir"`
	Moving bool        `json:"moving"`
}
