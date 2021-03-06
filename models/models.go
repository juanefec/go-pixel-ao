package models

import (
	"encoding/json"

	"github.com/segmentio/ksuid"
)

// Event type represents a message type where:
// 	-UpdateClient: client <- server
// 	-UpdateServer: client -> server
// 	-Spell:		   client <-> server
type Event int

// Events
const (
	UpdateClient Event = iota
	UpdateServer
	Spell
	Chat
	Death
	UpdateRanking
	ConfirmIDReception
	Disconect
)

func (d Event) String() string {
	return [...]string{"UpdateClient", "UpdateServer", "Spell", "Chat", "Death", "UpdateRanking", "Disconect"}[d]
}

type Mesg struct {
	Type    Event           `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

// NewMesg creates a new *Mesg
func NewMesg(t Event, payload json.RawMessage) []byte {
	m := &Mesg{
		Type:    t,
		Payload: payload,
	}
	msg, _ := json.Marshal(m)
	return msg
}

// UnmarshallMesg decodes incoming []byte into *Mesg
func UnmarshallMesg(m []byte) *Mesg {
	r := &Mesg{}
	json.Unmarshal(m, r)
	return r
}

type DisconectMsg struct {
	ID ksuid.KSUID `json:"id"`
}

type SpellMsg struct {
	ID         ksuid.KSUID `json:"id"`
	SpellType  string      `json:"spell_type"`
	SpellName  string      `json:"spell_name"`
	TargetID   ksuid.KSUID `json:"target_id"`
	Name       string      `json:"name"`
	X          float64     `json:"x"`
	Y          float64     `json:"y"`
	ChargeTime float64     `json:"charge_time"`
}

type PlayerMsg struct {
	ID        ksuid.KSUID `json:"id"`
	Name      string      `json:"name"`
	Skin      int         `json:"skin"`
	HP        float64     `json:"hp"`
	X         float64     `json:"x"`
	Y         float64     `json:"y"`
	Dir       string      `json:"dir"`
	Moving    bool        `json:"moving"`
	Dead      bool        `json:"dead"`
	Invisible bool        `json:"invisible"`
}

type ChatMsg struct {
	ID      ksuid.KSUID `json:"id"`
	Name    string      `json:"name"`
	Message string      `json:"message"`
}

type DeathMsg struct {
	Killed     ksuid.KSUID `json:"killed_id"`
	KilledName string      `json:"killed_name"`
	Killer     ksuid.KSUID `json:"killer_id"`
	KillerName string      `json:"killer_name"`
}

type RankingPosMsg struct {
	Name string      `json:"name"`
	ID   ksuid.KSUID `json:"id"`
	K    int         `json:"kills"`
	D    int         `json:"deaths"`
}
