package models

import "github.com/segmentio/ksuid"

type PlayerMessage struct {
	Apoca  ApocaMsg  `json:"apoca"`
	Player PlayerMsg `json:"player"`
}

type UpdateMessage struct {
	Apocas  []*ApocaMsg  `json:"apocas"`
	Players []*PlayerMsg `json:"players"`
}
type ApocaMsg struct {
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
