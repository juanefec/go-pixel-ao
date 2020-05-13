package main

import (
	"time"

	"github.com/faiface/pixel/pixelgl"
)

type MapBound float64

const (
	Top    = 4000
	Bottom = 0
	Left   = 0
	Right  = 4000
)

type KeyConfig struct {
	Apoca          int `json:"apoca_key"`
	Desca          int `json:"desca_key"`
	Explo          int `json:"explo_key"`
	PrimarySkill   int `json:"primary_skill_key"`
	SecondarySkill int `json:"secondary_skill_key"`
	Rojas          int `json:"rojas_key"`
	Azules         int `json:"azules_key"`
}

var min = 99999999999999

func keyInputs(win *pixelgl.Window, player *Player, cursor *Cursor, cs *CollisionSystem) { // collision system should be removed in a future refactor
	last := time.Now()
	const (
		KeyUp    = pixelgl.KeyW
		KeyDown  = pixelgl.KeyS
		KeyLeft  = pixelgl.KeyA
		KeyRight = pixelgl.KeyD
	)

	timeMap := map[pixelgl.Button]int{
		KeyUp:    -1,
		KeyDown:  -1,
		KeyLeft:  -1,
		KeyRight: -1,
	}

	latestPressed := func(keyPressed pixelgl.Button, m map[pixelgl.Button]int) bool {
		var key pixelgl.Button
		min := min
		for k, v := range m {
			if v < min && v >= 0 {
				key = k
				min = v
			}
		}
		return key == keyPressed
	}

	//tpTime := time.Now()

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()
		if !player.chat.chatting && !player.rooted {
			var axisX bool
			dist := .0
			if win.Pressed(KeyLeft) {
				if latestPressed(KeyLeft, timeMap) {
					player.moving = true
					player.dir = "left"
					if player.bounds.pos.X > Left {
						axisX = true
						dist -= player.playerMovementSpeed * dt
						timeMap[KeyLeft] = 0
					} else {
						player.moving = false
					}
				}
				timeMap[KeyLeft]++
			} else {
				timeMap[KeyLeft] = -1
			}

			if win.Pressed(KeyRight) {
				if latestPressed(KeyRight, timeMap) {
					player.moving = true
					player.dir = "right"
					if player.bounds.pos.X < Right {
						axisX = true
						dist += player.playerMovementSpeed * dt
						timeMap[KeyRight] = 0
					} else {
						player.moving = false
					}
				}
				timeMap[KeyRight]++
			} else {
				timeMap[KeyRight] = -1
			}

			if win.Pressed(KeyDown) {
				if latestPressed(KeyDown, timeMap) {
					player.moving = true
					player.dir = "down"
					if player.bounds.pos.Y > Bottom {
						axisX = false
						dist -= player.playerMovementSpeed * dt
						timeMap[KeyDown] = 0
					} else {
						player.moving = false
					}
				}
				timeMap[KeyDown]++

			} else {
				timeMap[KeyDown] = -1
			}

			if win.Pressed(KeyUp) {
				if latestPressed(KeyUp, timeMap) {
					player.moving = true
					player.dir = "up"
					if player.bounds.pos.Y < Top {
						axisX = false
						dist += player.playerMovementSpeed * dt
						timeMap[KeyUp] = 0
					} else {
						player.moving = false
					}
				}
				timeMap[KeyUp]++
			} else {
				timeMap[KeyUp] = -1
			}

			if player.moving {
				if axisX {
					player.bounds.pos.X += dist
					if len(cs.RetrieveIntersections(&player.bounds)) != 0 {
						player.bounds.pos.X -= dist
					}
				} else {
					player.bounds.pos.Y += dist
					if len(cs.RetrieveIntersections(&player.bounds)) != 0 {
						player.bounds.pos.Y -= dist
					}
				}
			}

			if win.JustPressed(pixelgl.Button(Key.Explo)) {
				cursor.SetSpellExploMode()
			}

			if win.JustPressed(pixelgl.Button(Key.Apoca)) {
				cursor.SetSpellApocaMode()
			}

			if win.JustPressed(pixelgl.Button(Key.Desca)) {
				cursor.SetSpellDescaMode()
			}

		} else {
			player.moving = false
		}
		if win.Pressed(pixelgl.Button(Key.Azules)) {
			player.drinkingManaPotions = true
		} else {
			player.drinkingManaPotions = false
		}

		if win.Pressed(pixelgl.Button(Key.Rojas)) {
			player.drinkingHealthPotions = true
		} else {
			player.drinkingHealthPotions = false
		}

		if timeMap[KeyUp] == -1 && timeMap[KeyDown] == -1 && timeMap[KeyLeft] == -1 && timeMap[KeyRight] == -1 {
			player.moving = false
		}
	}
}
