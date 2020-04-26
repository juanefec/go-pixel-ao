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

func keyInputs(win *pixelgl.Window, player *Player, cursor *Cursor) {
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
		min := 99999999999999
		for k, v := range m {
			if v < min && v > 0 {
				key = k
				min = v
			}
		}
		return key == keyPressed
	}

	for !win.Closed() {
		dt := time.Since(last).Seconds()
		last = time.Now()

		if win.Pressed(KeyLeft) {
			if latestPressed(KeyLeft, timeMap) {
				player.moving = true
				player.dir = "left"
				if player.pos.X > Left {
					player.pos.X -= PlayerSpeed * dt
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
				if player.pos.X < Right {
					player.pos.X += PlayerSpeed * dt
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
				if player.pos.Y > Bottom {
					player.pos.Y -= PlayerSpeed * dt
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
				if player.pos.Y < Top {
					player.pos.Y += PlayerSpeed * dt
				} else {
					player.moving = false
				}
			}
			timeMap[KeyUp]++
		} else {
			timeMap[KeyUp] = -1
		}

		if win.Pressed(pixelgl.KeyF) {
			player.drinkingManaPotions = true
		} else {
			player.drinkingManaPotions = false
		}

		if win.Pressed(pixelgl.MouseButtonRight) {
			player.drinkingHealthPotions = true
		} else {
			player.drinkingHealthPotions = false
		}

		if win.JustPressed(pixelgl.Key2) {
			cursor.SetSpellApocaMode()
		}

		if win.JustPressed(pixelgl.Key3) {
			cursor.SetSpellDescaMode()
		}

		if player.sname == "creagod" && win.Pressed(pixelgl.KeyLeftShift) {
			if win.JustPressed(pixelgl.MouseButtonLeft) {
				tppos := player.cam.Unproject(win.MousePosition())
				player.pos.X, player.pos.Y = tppos.X, tppos.Y
			}
		}

		if timeMap[KeyUp] == -1 && timeMap[KeyDown] == -1 && timeMap[KeyLeft] == -1 && timeMap[KeyRight] == -1 {
			player.moving = false
		}
	}
}
