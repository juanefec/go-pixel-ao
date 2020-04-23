package main

import (
	"time"

	"github.com/faiface/pixel/pixelgl"
)

func keyInputs(win *pixelgl.Window, player *Player) {
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
				player.pos.X -= PlayerSpeed * dt
			}
			timeMap[KeyLeft]++
		} else {
			timeMap[KeyLeft] = -1
		}

		if win.Pressed(KeyRight) {
			if latestPressed(KeyRight, timeMap) {
				player.moving = true
				player.dir = "right"
				player.pos.X += PlayerSpeed * dt
			}
			timeMap[KeyRight]++
		} else {
			timeMap[KeyRight] = -1
		}

		if win.Pressed(KeyDown) {
			if latestPressed(KeyDown, timeMap) {
				player.moving = true
				player.dir = "down"
				player.pos.Y -= PlayerSpeed * dt
			}
			timeMap[KeyDown]++

		} else {
			timeMap[KeyDown] = -1
		}

		if win.Pressed(KeyUp) {
			if latestPressed(KeyUp, timeMap) {
				player.moving = true
				player.dir = "up"
				player.pos.Y += PlayerSpeed * dt
			}
			timeMap[KeyUp]++
		} else {
			timeMap[KeyUp] = -1
		}

		if timeMap[KeyUp] == -1 && timeMap[KeyDown] == -1 && timeMap[KeyLeft] == -1 && timeMap[KeyRight] == -1 {
			player.moving = false
		}
	}
}
